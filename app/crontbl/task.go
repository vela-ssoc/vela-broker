package crontbl

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

type cronTask struct {
	ctx  context.Context
	idle time.Duration
	log  *slog.Logger
	fn   func(ctx context.Context, at time.Time) error
}

func (ct *cronTask) Run() {
	ticker := time.NewTicker(ct.idle)
	defer ticker.Stop()

	var over bool
	for !over {
		select {
		case <-ct.ctx.Done():
			over = true
		case at := <-ticker.C:
			if err := ct.exec(at); err != nil {
				ct.log.Warn("cron 任务执行错误", slog.Any("error", err))
			}
		}
	}
}

func (ct *cronTask) exec(at time.Time) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("cron exec panic: %v", v)
		}
	}()

	ctx := ct.ctx
	err = ct.fn(ctx, at)

	return err
}
