package launch

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/vela-ssoc/broker/brkcli"
	"github.com/vela-ssoc/broker/central"
	"github.com/vela-ssoc/broker/infra/bootstrap"
	"github.com/vela-ssoc/broker/infra/logback"
)

func Run(parent context.Context, cfg string, logger logback.Replacer) error {
	var hide brkcli.Hide
	if err := bootstrap.AutoLoad(cfg, os.Args[0], &hide); err != nil {
		return err
	}

	brk, err := brkcli.MustJoin(parent, hide, logger)
	if err != nil {
		return err
	}

	center := central.New()

	for {
		ctx, cancel := context.WithCancel(parent)
		srv := &http.Server{Handler: center}
		lis := brk.Listener()

		go watch(ctx, brk, srv)

		_ = srv.Serve(lis)
		cancel()

		if cex := parent.Err(); cex != nil {
			return cex
		}
		if _ = brk.Reconnect(parent); err != nil {
			return err
		}
	}
}

func watch(ctx context.Context, brk brkcli.Broker, srv *http.Server) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var failed int

	for {
		select {
		case <-ctx.Done():
			_ = srv.Close()
			return
		case <-ticker.C:
			if err := brk.FetchNonReturn(ctx, brkcli.BrkPing, nil); err == nil {
				failed = 0
			} else {
				if failed++; failed > 3 {
					_ = srv.Close()
					return
				}
			}
		}
	}
}
