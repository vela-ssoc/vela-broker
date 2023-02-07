package launch

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/vela-ssoc/broker/brkcli"
	"github.com/vela-ssoc/broker/infra/bootstrap"
	"github.com/vela-ssoc/broker/infra/logback"
)

func Run(ctx context.Context, cfg string, logger logback.Replacer) error {
	var hide brkcli.Hide
	if err := bootstrap.AutoLoad(cfg, os.Args[0], &hide); err != nil {
		return err
	}

	brk, err := brkcli.MustJoin(ctx, hide, logger)
	if err != nil {
		return err
	}

	for {
		heartbeat(ctx, brk)
		if err = brk.Reconnect(ctx); err != nil {
			return err
		}
	}
}

func heartbeat(ctx context.Context, brk brkcli.Broker) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := brk.FetchNonReturn(ctx, brkcli.BrkPing, nil)
			log.Println(err)
		}
	}
}
