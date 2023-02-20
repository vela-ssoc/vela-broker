package brkcli

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/dfcfw/spdy"
	"github.com/vela-ssoc/backend-common/httpclient"
	"github.com/vela-ssoc/vela-broker/infra/logback"
)

var (
	ErrRequiredID     = errors.New("id 不能为空")
	ErrRequiredSecret = errors.New("secret 不能为空")
	ErrRequiredSemver = errors.New("semver 不能为空")
	ErrRequiredServer = errors.New("servers 不能为空")
)

type Broker interface {
	// Hide 信息
	Hide() Hide

	// Ident 身份认证信息
	Ident() Ident

	// Issue 中心端下发的配置
	Issue() Issue

	// Listener 获取 net.Listener
	Listener() net.Listener

	// Reconnect 重连
	Reconnect(context.Context) error

	// Fetch 发送请求
	Fetch(context.Context, Operator, io.Reader) (*http.Response, error)

	// PostJSON 向中心端发送 JSON 数据，返回 JSON 数据
	PostJSON(context.Context, Operator, any, any) error

	Oneway(context.Context, Operator, io.Reader) error
}

// MustJoin 接入
func MustJoin(ctx context.Context, hide Hide, slog logback.Logger) (Broker, error) {
	if hide.ID == 0 {
		return nil, ErrRequiredID
	}
	if hide.Semver == "" {
		return nil, ErrRequiredSemver
	}
	if hide.Secret == "" {
		return nil, ErrRequiredSecret
	}
	if len(hide.Servers) == 0 {
		return nil, ErrRequiredServer
	}
	hide.Servers.Format()

	dial := newMustDial(hide.Servers, slog)
	bn := &brokerNode{
		hide:      hide,
		dialer:    dial,
		logger:    slog,
		userAgent: "Broker-Client/" + hide.Semver,
	}

	transport := &http.Transport{DialContext: bn.dialContext}
	under := &http.Client{Transport: transport}
	bn.client = httpclient.NewClient(under)

	if err := bn.mustJoin(ctx); err != nil {
		return nil, err
	}

	return bn, nil
}

// brokerNode broker 节点客户端
type brokerNode struct {
	hide      Hide               // 启动配置信息
	ident     Ident              // 认证信息
	issue     Issue              // 认证返回信息
	dialer    mustDialer         // TCP 连接器
	logger    logback.Logger     // 日志输出
	client    httpclient.Client  // HTTP Client
	mux       spdy.Muxer         // TCP 多路复用
	userAgent string             // HTTP User-Agent
	ctx       context.Context    // context.Context
	cancel    context.CancelFunc // context.CancelFunc
}

func (bn *brokerNode) Hide() Hide {
	return bn.hide
}

func (bn *brokerNode) Ident() Ident {
	return bn.ident
}

func (bn *brokerNode) Issue() Issue {
	return bn.issue
}

func (bn *brokerNode) Listener() net.Listener {
	return bn.mux
}

func (bn *brokerNode) Reconnect(ctx context.Context) error {
	bn.cancel()
	_ = bn.mux.Close()

	return bn.mustJoin(ctx)
}

func (bn *brokerNode) Fetch(ctx context.Context, op Operator, body io.Reader) (*http.Response, error) {
	req := bn.newRequest(ctx, op, body)
	return bn.fetch(req)
}

func (bn *brokerNode) Oneway(ctx context.Context, op Operator, body io.Reader) error {
	req := bn.newRequest(ctx, op, body)
	res, err := bn.fetch(req)
	if err == nil {
		_ = res.Body.Close()
	}
	return err
}

func (bn *brokerNode) PostJSON(ctx context.Context, op Operator, body, ret any) error {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(body); err != nil {
		return err
	}

	req := bn.newRequest(ctx, op, buf)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	res, err := bn.fetch(req)
	if err != nil {
		return err
	}
	err = json.NewDecoder(res.Body).Decode(ret)
	_ = res.Body.Close()

	return err
}

func (bn *brokerNode) fetch(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", bn.userAgent)
	return bn.client.Fetch(req)
}

func (bn *brokerNode) newRequest(ctx context.Context, op Operator, body io.Reader) *http.Request {
	method := op.Method()
	path := op.Path()
	// Host 名字没有意义，但是如果不设置 Host ，http 标准库会检查报错
	addr := &url.URL{Scheme: "http", Host: "soc", Path: path}
	req := &http.Request{
		Method:     method,
		URL:        addr,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if v, ok := body.(interface{ Len() int }); ok {
		req.ContentLength = int64(v.Len())
	}
	switch v := body.(type) {
	case nil:
	case io.ReadCloser:
		req.Body = v
	case *bytes.Buffer:
		req.Body = io.NopCloser(v)
	case *bytes.Reader:
		req.Body = io.NopCloser(v)
	case *strings.Reader:
		req.Body = io.NopCloser(v)
	default:
		req.ContentLength = -1
		req.Body = io.NopCloser(body)
	}

	// For client requests, Request.ContentLength of 0
	// means either actually 0, or unknown. The only way
	// to explicitly say that the ContentLength is zero is
	// to set the Body to nil. But turns out too much code
	// depends on NewRequest returning a non-nil Body,
	// so we use a well-known ReadCloser variable instead
	// and have the http package also treat that sentinel
	// variable to mean explicitly zero.
	if req.Body != nil && req.ContentLength == 0 {
		req.Body = http.NoBody
	}

	if ctx == nil {
		ctx = context.Background()
	}

	return req.WithContext(ctx)
}

func (bn *brokerNode) dialContext(_ context.Context, _, _ string) (net.Conn, error) {
	mux := bn.mux
	if mux == nil {
		return nil, net.ErrClosed
	}

	return mux.Dial()
}

func (bn *brokerNode) mustJoin(ctx context.Context) error {
	bn.ctx, bn.cancel = context.WithCancel(ctx)
	sleep := time.Second
	for {
		conn, srv, err := bn.dialer.MustDial(ctx, 3*time.Second)
		if err != nil {
			return err
		}
		if err = bn.handshake(ctx, conn, srv); err == nil {
			return nil
		}
		_ = conn.Close()
		bn.logger.Infof("协议认证失败 %s 后重试：%s %v", sleep, srv, err)

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}
	}
}

func (bn *brokerNode) handshake(ctx context.Context, conn net.Conn, addr *Address) error {
	ip := conn.LocalAddr().(*net.TCPAddr).IP
	mac := bn.dialer.LookupMAC(ip)

	ident := Ident{
		ID:     bn.hide.ID,
		Secret: bn.hide.Secret,
		IP:     ip,
		MAC:    mac.String(),
		Semver: bn.hide.Semver,
		Goos:   runtime.GOOS,
		Arch:   runtime.GOARCH,
		TimeAt: time.Now(),
	}

	hostname, _ := os.Hostname()
	ident.Hostname = hostname
	pid := os.Getpid()
	ident.PID = pid
	workdir, _ := os.Getwd()
	ident.Workdir = workdir
	executable, _ := os.Executable()
	ident.Executable = executable
	cpu := runtime.NumCPU()
	ident.CPU = cpu
	if cu, _ := user.Current(); cu != nil {
		ident.Username = cu.Username
	}

	data, err := ident.Encrypt()
	if err != nil {
		return err
	}
	buf := bytes.NewReader(data)

	host := addr.Name
	if host == "" {
		dest := addr.Addr
		if idx := strings.LastIndex(dest, ":"); idx > -1 {
			host = dest[:idx]
		}
	}

	req := bn.newRequest(ctx, opJoin, buf)
	req.Host = host
	req.URL.Host = host
	if err = req.Write(conn); err != nil {
		return err
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer res.Body.Close()

	code := res.StatusCode
	if code != http.StatusAccepted {
		cause := make([]byte, 4096)
		n, _ := io.ReadFull(res.Body, cause)
		ret := struct {
			Message string `json:"message"`
		}{}
		exr := &httpclient.Error{Code: code}
		if err = json.Unmarshal(cause[:n], &ret); err == nil {
			exr.Text = []byte(ret.Message)
		} else {
			exr.Text = cause[:n]
		}

		return exr
	}

	enc := make([]byte, 40960)
	n, _ := res.Body.Read(enc)
	var issue Issue
	if err = issue.Decrypt(enc[:n]); err != nil {
		return err
	}

	bn.ident = ident
	bn.issue = issue
	bn.mux = spdy.Client(conn, spdy.WithEncrypt(issue.Passwd))

	return nil
}
