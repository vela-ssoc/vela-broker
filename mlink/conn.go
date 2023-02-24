package mlink

import (
	"context"
	"net"

	"github.com/vela-ssoc/backend-common/spdy"
)

type Infer interface {
	Ident() Ident
	Issue() Issue
	Inet() net.IP
}

type connect struct {
	id    int64
	ident Ident
	issue Issue
	mux   spdy.Muxer
}

func (c *connect) Ident() Ident { return c.ident }
func (c *connect) Issue() Issue { return c.issue }
func (c *connect) Inet() net.IP { return c.ident.Inet }

type contextKey struct{ name string }

var minionCtxKey = &contextKey{name: "minion-context"}

func Ctx(ctx context.Context) Infer {
	if ctx != nil {
		infer, _ := ctx.Value(minionCtxKey).(Infer)
		return infer
	}

	return nil
}
