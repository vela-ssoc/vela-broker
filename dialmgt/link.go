package dialmgt

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"

	"github.com/vela-ssoc/backend-common/httpclient"
	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/opurl"
)

var ErrEmptyAddress = errors.New("服务端地址不能为空")

type Linker interface {
	Hide() Hide
	Ident() Ident
	Issue() Issue
	Oneway(context.Context, opurl.URLer, io.Reader) error
	Call(context.Context, opurl.URLer, io.Reader) (*http.Response, error)
	Listen() net.Listener
	NodeName() string
	Reconnect(context.Context) error
}

func Dial(parent context.Context, hide Hide, slog logback.Logger) (Linker, error) {
	if len(hide.Servers) == 0 {
		return nil, ErrEmptyAddress
	}
	hide.Servers.Format()

	dialer := newIterDial(hide.Servers)

	bc := &brokerClient{
		hide:   hide,
		slog:   slog,
		dialer: dialer,
	}
	transport := &http.Transport{DialContext: bc.dialContext}
	cli := &http.Client{Transport: transport}
	bc.client = httpclient.NewClient(cli)

	if err := bc.dial(parent); err != nil {
		return nil, err
	}

	// go bc.heartbeat(time.Minute)

	return bc, nil
}
