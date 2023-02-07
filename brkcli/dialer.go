package brkcli

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/vela-ssoc/broker/infra/logback"
)

type mustDialer interface {
	MustDial(ctx context.Context, timeout time.Duration, sleep time.Duration) (net.Conn, *Server, error)
	LookupMAC(net.IP) net.HardwareAddr
}

func newMustDial(servers []*Server, logger logback.Logger) mustDialer {
	macs := make(map[string]net.HardwareAddr, len(servers))
	dial := &tls.Dialer{NetDialer: new(net.Dialer)}

	return &ringDial{
		logger: logger,
		dialer: dial,
		macs:   macs,
		dest:   servers,
	}
}

type ringDial struct {
	logger logback.Logger
	dialer *tls.Dialer
	macs   map[string]net.HardwareAddr
	dest   []*Server
	index  int
}

func (dl *ringDial) MustDial(ctx context.Context, timeout, sleep time.Duration) (net.Conn, *Server, error) {
	for {
		srv := dl.dest[dl.index]
		conn, err := dl.dial(ctx, srv, timeout)
		dl.index = (dl.index + 1) % len(dl.dest)
		if err == nil {
			return conn, srv, nil
		}

		dl.logger.Infof("连接服务器 [%s] 失败，%s 后重新连接：%v", srv, sleep, err)

		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		case <-time.After(sleep):
		}
	}
}

// LookupMAC 查询 net.IP 所在网卡的 MAC 地址，非并发安全
func (dl *ringDial) LookupMAC(ip net.IP) net.HardwareAddr {
	sip := ip.String()
	if hw, ok := dl.macs[sip]; ok {
		return hw
	}

	var mac net.HardwareAddr
	ifs, _ := net.Interfaces()
	for _, face := range ifs {

		var match bool
		addrs, _ := face.Addrs()
		for _, addr := range addrs {
			inet, ok := addr.(*net.IPNet)
			if match = ok && inet.IP.Equal(ip); match {
				break
			}
		}
		if match {
			mac = face.HardwareAddr
			break
		}
	}

	dl.macs[sip] = mac

	return mac
}

func (dl *ringDial) dial(parent context.Context, srv *Server, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	if srv.TLS {
		dl.dialer.Config = &tls.Config{ServerName: srv.Name}
		return dl.dialer.DialContext(ctx, "tcp", srv.Addr)
	} else {
		return dl.dialer.NetDialer.DialContext(ctx, "tcp", srv.Addr)
	}
}
