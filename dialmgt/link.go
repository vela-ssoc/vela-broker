package dialmgt

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/vela-ssoc/backend-common/logback"
)

var ErrEmptyAddress = errors.New("服务端地址不能为空")

type Linker interface {
	Hide() Hide
	Ident() Ident
	Issue() Issue
	Listen() net.Listener
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
	if err := bc.dial(parent); err != nil {
		return nil, err
	}

	go bc.heartbeat(5 * time.Second)

	return bc, nil
}
