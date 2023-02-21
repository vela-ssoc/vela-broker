package brkcli

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"github.com/vela-ssoc/backend-common/logback"
)

type mustDialer interface {
	MustDial(ctx context.Context, timeout time.Duration) (net.Conn, *Address, error)
	LookupMAC(net.IP) net.HardwareAddr
}

func newMustDial(servers Addresses, slog logback.Logger) mustDialer {
	macs := make(map[string]net.HardwareAddr, len(servers))
	dial := &tls.Dialer{NetDialer: new(net.Dialer)}

	return &ringDial{
		logger: slog,
		dialer: dial,
		macs:   macs,
		dest:   servers,
	}
}

type ringDial struct {
	logger logback.Logger
	dialer *tls.Dialer
	macs   map[string]net.HardwareAddr
	dest   Addresses
	index  int
}

func (dl *ringDial) MustDial(ctx context.Context, timeout time.Duration) (net.Conn, *Address, error) {
	begin := time.Now()
	dsz := len(dl.dest)
	for {
		srv := dl.dest[dl.index]
		dl.index = (dl.index + 1) % dsz
		conn, err := dl.dial(ctx, srv, timeout)
		if err == nil {
			return conn, srv, nil
		}

		sleep := dl.sleepN(begin)
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

func (dl *ringDial) dial(parent context.Context, srv *Address, timeout time.Duration) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	if srv.TLS {
		dl.dialer.Config = &tls.Config{ServerName: srv.Name}
		return dl.dialer.DialContext(ctx, "tcp", srv.Addr)
	} else {
		return dl.dialer.NetDialer.DialContext(ctx, "tcp", srv.Addr)
	}
}

func (*ringDial) sleepN(begin time.Time) time.Duration {
	since := time.Since(begin)
	switch {
	case since > time.Hour:
		return 10 * time.Second
	case since > time.Minute:
		return 3 * time.Second
	default:
		return time.Second
	}
}
