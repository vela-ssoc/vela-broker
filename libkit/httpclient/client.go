package httpclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	cli *http.Client
}

func NewClient(cli ...*http.Client) Client {
	if len(cli) != 0 {
		return Client{cli: cli[0]}
	}

	hc := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   true,
			DisableCompression:  true,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
			MaxIdleConns:        10,
			MaxConnsPerHost:     10,
			MaxIdleConnsPerHost: 10,
		},
	}

	return Client{cli: hc}
}

// Get 发送 GET 请求
func (hc Client) Get(ctx context.Context, addr string, queries url.Values, opts ...Option) (*http.Response, error) {
	return hc.Request(ctx, http.MethodGet, addr, queries, nil, opts...)
}

// GetJSON 发送 GET 请求，返回数据为 JSON
func (hc Client) GetJSON(ctx context.Context, addr string, queries url.Values, ret any, opts ...Option) error {
	opt := WithHeader("Accept", "application/json")
	opts = append(opts, opt)

	res, err := hc.Get(ctx, addr, queries, opts...)
	if err == nil {
		err = json.NewDecoder(res.Body).Decode(ret)
		_ = res.Body.Close()
	}

	return err
}

// Post 发送 POST 请求
func (hc Client) Post(ctx context.Context, addr string, queries url.Values, body io.Reader, opts ...Option) (*http.Response, error) {
	return hc.Request(ctx, http.MethodPost, addr, queries, body, opts...)
}

// PostForm 发送 POST 请求，数据 body 为 FormData 格式
func (hc Client) PostForm(ctx context.Context, addr string, queries url.Values, body url.Values, opts ...Option) (*http.Response, error) {
	opt := WithHeader("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	opts = append(opts, opt)
	br := strings.NewReader(body.Encode())

	return hc.Post(ctx, addr, queries, br, opts...)
}

// PostJSON 发送 POST 请求，请求格式为 JSON，返回数据格式为 JSON
func (hc Client) PostJSON(ctx context.Context, addr string, queries url.Values, body, ret any, opts ...Option) error {
	buf := new(bytes.Buffer)
	if body != nil {
		if err := json.NewEncoder(buf).Encode(body); err != nil {
			return err
		}
	}

	accept := WithHeader("Accept", "application/json")
	ct := WithHeader("Content-Type", "application/json; charset=utf-8")
	opts = append(opts, accept, ct)

	res, err := hc.Post(ctx, addr, queries, buf, opts...)
	// 请求成功但是没有响应报文，比如：http.StatusNoContent
	if err == nil {
		err = json.NewDecoder(res.Body).Decode(ret)
		_ = res.Body.Close()
	}

	return err
}

// Put 发送 HTTP PUT 请求
func (hc Client) Put(ctx context.Context, addr string, queries url.Values, body io.Reader, opts ...Option) (*http.Response, error) {
	return hc.Request(ctx, http.MethodPut, addr, queries, body, opts...)
}

func (hc Client) Request(
	ctx context.Context,
	method string,
	addr string,
	queries url.Values,
	body io.Reader,
	opts ...Option,
) (res *http.Response, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if addr, err = mergeQuery(addr, queries); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, addr, body)
	if err != nil {
		return nil, err
	}

	return hc.Fetch(req, opts...)
}

func (hc Client) Fetch(req *http.Request, opts ...Option) (res *http.Response, err error) {
	opt := &option{header: http.Header{}}
	opt.merge(opts...)

	if opt.host != "" {
		req.Host = opt.host
	}
	for k, vs := range opt.header {
		for _, v := range vs {
			req.Header.Add(k, v)
		}
	}

	res, err = hc.send(req)
	if err == nil || opt.retry <= 0 || !hc.canRetry(err) {
		return res, err
	}
	for i := 0; i < opt.retry; i++ {
		time.Sleep(opt.delay)
		if res, err = hc.send(req); err == nil || !hc.canRetry(err) {
			break
		}
	}

	return
}

// send 发送 HTTP 请求
func (hc Client) send(req *http.Request) (*http.Response, error) {
	res, err := hc.cli.Do(req)
	if err != nil {
		return nil, err
	}

	code := res.StatusCode
	if code >= http.StatusOK && code < http.StatusMultipleChoices {
		return res, nil
	}

	buf := make([]byte, 1024)
	n, _ := io.ReadFull(res.Body, buf)
	_ = res.Body.Close()
	err = &Error{Code: code, Text: buf[:n]}

	return nil, err
}

// retry 判断是否需要重试请求
func (Client) canRetry(err error) bool {
	switch e := err.(type) {
	case *Error:
		return e.Code == http.StatusTooManyRequests || e.Code >= http.StatusInternalServerError
	default:
		return false
	}
}

// appendQueries 将参数合并
//
//	example:
//			addr: https://18.com.cn/?name=jack
//			queries: map[age][]string{"18"}
//		合并后: https://18.com.cn/?name=jack&age=18
func mergeQuery(addr string, queries url.Values) (string, error) {
	if len(queries) == 0 {
		return addr, nil
	}

	u, err := url.Parse(addr)
	if err != nil {
		return "", err
	}
	if u.RawQuery != "" {
		values, ex := url.ParseQuery(u.RawQuery)
		if ex != nil {
			return "", ex
		}
		for key, vals := range values {
			for _, val := range vals {
				queries.Add(key, val)
			}
		}
	}

	u.RawQuery = queries.Encode()

	return u.String(), nil
}
