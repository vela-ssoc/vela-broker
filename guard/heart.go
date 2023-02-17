package guard

import (
	"context"
	"time"

	"github.com/vela-ssoc/broker/brkcli"
)

func newHeart(ctx context.Context, brk brkcli.Broker, du time.Duration) *heart {
	return &heart{
		du:  du,
		brk: brk,
		ctx: ctx,
	}
}

type heart struct {
	du  time.Duration
	brk brkcli.Broker
	ctx context.Context
}

func (h *heart) Run() {
	ticker := time.NewTicker(h.du)
	defer ticker.Stop()

	for {
		select {
		case <-h.ctx.Done():
			return
		case <-ticker.C:
			_ = h.brk.Oneway(nil, brkcli.BrkPing, nil)
		}
	}
}
