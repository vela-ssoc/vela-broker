package netdial

import (
	"context"
	"crypto/tls"
	"net"
	"time"
)

type Dialer interface {
	DialContext(context.Context, time.Duration) (net.Conn, error)
}

func Hostname(addr, name string) Dialer {
	if name == "" {
		if host, _, _ := net.SplitHostPort(addr); host != "" {
			name = host
		}
	}
	dial := &tls.Dialer{
		NetDialer: new(net.Dialer),
		Config:    &tls.Config{ServerName: name},
	}

	return &hostname{
		addr: addr,
		name: name,
		dial: dial,
	}
}

type hostname struct {
	addr string
	name string
	dial *tls.Dialer
}

func (hn *hostname) DialContext(parent context.Context, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	conn, err := hn.dial.NetDialer.DialContext(ctx, "tcp", hn.addr)
	if err != nil {
		return nil, err
	}
	ssl := tls.Client(conn, hn.dial.Config)
	err = ssl.HandshakeContext(ctx)
	if err == nil {
		return ssl, nil
	}
	if err == context.DeadlineExceeded {
		return conn, nil
	}
	if _, ok := err.(tls.RecordHeaderError); ok {
		return conn, nil
	}
	_ = ssl.Close()

	return nil, err
}
