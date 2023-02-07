package httpclient

import (
	"encoding/base64"
	"net/http"
	"time"
)

type option struct {
	header  http.Header   // 附加的 header
	timeout time.Duration // 请求超时时间
	retry   int           // 如果失败，重试次数
	delay   time.Duration // 每次重试的时间间隔
	host    string        // 自定义的 Host
}

type Option func(o *option)

func WithHeader(key, val string) Option {
	return func(o *option) {
		o.header.Add(key, val)
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(o *option) {
		o.timeout = timeout
	}
}

func WithRetry(n int) Option {
	return func(o *option) {
		o.retry = n
	}
}

func WithBasicAuth(username, password string) Option {
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	val := "Basic " + auth
	return func(o *option) {
		o.header.Add("Authorization", val)
	}
}

func WithDelay(delay time.Duration) Option {
	return func(o *option) {
		o.delay = delay
	}
}

func WithHost(host string) Option {
	return func(o *option) {
		o.host = host
	}
}

func (opt *option) merge(opts ...Option) {
	for _, optFn := range opts {
		optFn(opt)
	}
	if opt.timeout <= 0 {
		opt.timeout = 10 * time.Second
	}
	if opt.retry > 0 && opt.delay <= 0 {
		opt.delay = 200 * time.Millisecond
	}
}
