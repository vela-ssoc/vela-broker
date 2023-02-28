package dialmgt

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/user"
	"runtime"
	"strings"
	"time"

	"github.com/vela-ssoc/backend-common/httpclient"
	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/opurl"
	"github.com/vela-ssoc/backend-common/spdy"
)

type brokerClient struct {
	hide   Hide
	ident  Ident
	issue  Issue
	slog   logback.Logger
	client httpclient.Client
	dialer *iterDial
	mux    spdy.Muxer
	parent context.Context
	ctx    context.Context
	cancel context.CancelFunc
}

func (bc *brokerClient) Hide() Hide           { return bc.hide }
func (bc *brokerClient) Ident() Ident         { return bc.ident }
func (bc *brokerClient) Issue() Issue         { return bc.issue }
func (bc *brokerClient) Listen() net.Listener { return bc.mux }

func (bc *brokerClient) NodeName() string {
	return fmt.Sprintf("broker-%s-%d", bc.ident.Inet, bc.ident.ID)
}

func (bc *brokerClient) Oneway(ctx context.Context, op opurl.URLer, body io.Reader) error {
	res, err := bc.Call(ctx, op, body)
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	return nil
}

func (bc *brokerClient) Call(ctx context.Context, op opurl.URLer, body io.Reader) (*http.Response, error) {
	req := bc.newRequest(ctx, op, body)
	return bc.client.Fetch(req)
}

func (bc *brokerClient) Reconnect(parent context.Context) error {
	_ = bc.close()
	return bc.dial(parent)
}

func (bc *brokerClient) close() error {
	bc.cancel()
	return bc.mux.Close()
}

func (bc *brokerClient) dial(parent context.Context) error {
	bc.parent = parent
	bc.ctx, bc.cancel = context.WithCancel(parent)
	start := time.Now()

	for {
		conn, addr, err := bc.dialer.iterDial(bc.ctx, 5*time.Second)
		if err != nil {
			if ce := bc.ctx.Err(); ce != nil {
				return ce
			}
			bc.slog.Warnf("dial %s 失败：%s", addr, err)
			bc.dialSleep(bc.ctx, start)
			continue
		}

		ident, issue, err := bc.consult(bc.ctx, conn, addr)
		if err == nil {
			mux := spdy.Client(conn, spdy.WithEncrypt(issue.Passwd))
			bc.ident, bc.issue, bc.mux = ident, issue, mux
			return nil
		}

		_ = conn.Close()
		if pe := parent.Err(); pe != nil {
			return pe
		}
		if he, ok := err.(*httpclient.Error); ok && he.NotAcceptable() {
			return err
		}

		bc.slog.Warnf("与 %s 协商失败：%s", addr, err)
		bc.dialSleep(parent, start)
	}
}

// consult 当建立好 TCP 连接后进行应用层协商
func (bc *brokerClient) consult(parent context.Context, conn net.Conn, addr *Address) (Ident, Issue, error) {
	ip := conn.LocalAddr().(*net.TCPAddr).IP
	mac := bc.dialer.lookupMAC(ip)

	ident := Ident{
		ID:     bc.hide.ID,
		Secret: bc.hide.Secret,
		Semver: bc.hide.Semver,
		Inet:   ip,
		MAC:    mac.String(),
		Goos:   runtime.GOOS,
		Arch:   runtime.GOARCH,
		TimeAt: time.Now(),
	}
	ident.Hostname, _ = os.Hostname()
	ident.PID = os.Getpid()
	ident.Workdir, _ = os.Getwd()
	executable, _ := os.Executable()
	ident.Executable = executable
	ident.CPU = runtime.NumCPU()
	if cu, _ := user.Current(); cu != nil {
		ident.Username = cu.Username
	}

	var issue Issue
	enc, err := ident.Encrypt()
	if err != nil {
		return ident, issue, err
	}
	buf := bytes.NewReader(enc)

	host := addr.Name
	if host == "" {
		dest := addr.Addr
		if idx := strings.LastIndex(dest, ":"); idx > -1 {
			host = dest[:idx]
		}
	}

	req := bc.newRequest(parent, opurl.BrkJoin, buf)
	req.Host = host
	req.URL.Host = host
	if err = req.Write(conn); err != nil {
		return ident, issue, err
	}

	res, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return ident, issue, err
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

		return ident, issue, exr
	}

	resp := make([]byte, 40960)
	n, _ := res.Body.Read(resp)
	err = issue.Decrypt(resp[:n])

	return ident, issue, err
}

func (bc *brokerClient) dialSleep(ctx context.Context, start time.Time) {
	since := time.Since(start)
	du := time.Second

	switch {
	case since > 12*time.Hour:
		du = 10 * time.Minute
	case since > time.Hour:
		du = time.Minute
	case since > 30*time.Minute:
		du = 30 * time.Second
	case since > 10*time.Minute:
		du = 10 * time.Second
	case since > 3*time.Minute:
		du = 3 * time.Second
	}

	log.Printf("%s 后进行重试", du)
	// 非阻塞休眠
	select {
	case <-ctx.Done():
	case <-time.After(du):
	}
}

func (bc *brokerClient) newRequest(ctx context.Context, op opurl.URLer, body io.Reader) *http.Request {
	method := op.Method()
	addr := op.URL()
	req := &http.Request{
		Method:     method,
		URL:        addr,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
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
	if v, ok := body.(interface{ Len() int }); ok {
		req.ContentLength = int64(v.Len())
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

func (bc *brokerClient) dialContext(_ context.Context, _, _ string) (net.Conn, error) {
	if mux := bc.mux; mux != nil {
		return mux.Dial()
	}
	return nil, io.ErrNoProgress
}

func (bc *brokerClient) heartbeat(du time.Duration) {
	ticker := time.NewTicker(du)
	defer ticker.Stop()

over:
	for {
		select {
		case <-bc.parent.Done():
			break over
		case <-ticker.C:
			if err := bc.Oneway(nil, opurl.OpPing, nil); err != nil {
				bc.slog.Warnf("向 manager 发送心跳发生错误: %s", err)
			}
		}
	}
}
