package mlink

import (
	"context"

	"github.com/vela-ssoc/backend-common/spdy"
)

type Infer interface {
	Ident() Ident
	Issue() Issue
}

type connect struct {
	ident Ident
	issue Issue
	mux   spdy.Muxer
}

func (c *connect) Ident() Ident { return c.ident }
func (c *connect) Issue() Issue { return c.issue }

type contextKey struct{ name string }

var minionCtxKey = &contextKey{name: "minion-context"}

func UnwarpCtx(ctx context.Context) Infer {
	if ctx != nil {
		infer, _ := ctx.Value(minionCtxKey).(Infer)
		return infer
	}

	return nil
}
