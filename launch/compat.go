package launch

import (
	"context"
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/vela-ssoc/vela-broker/restapi/agtsrv"
	"github.com/vela-ssoc/vela-broker/restapi/listen"
	"github.com/vela-ssoc/vela-broker/restapi/mgtcli"
	"github.com/vela-ssoc/vela-broker/telmgt"
	"github.com/xgfone/ship/v5"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// CRun 与老的 broker 兼容性接口，升级后会删除。
//
// Deprecated: use Run
func CRun(ctx context.Context, hide telmgt.Hide, sh *ship.Ship, slog logback.Logger, ech chan<- error) {
	link, err := telmgt.Dial(ctx, hide, slog)
	if err != nil {
		ech <- err
		return
	}
	ident, issue := link.Ident(), link.Issue()
	zlg := issue.Logger.Zap()
	slog.Replace(zlg.WithOptions(zap.AddCallerSkip(1)))
	slog.Infof("broker 接入认证成功，上报认证信息如下：\n%s\n下发的配置如下：\n%s", ident, issue)
	node := link.NodeName()

	dfg := issue.Database
	lvl, dsn := dfg.Level, dfg.FormatDSN()
	dlg := logback.Gorm(zlg, lvl)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: dlg})
	if err != nil {
		ech <- err
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		ech <- err
		return
	}
	sqlDB.SetMaxIdleConns(dfg.MaxIdleConn)
	sqlDB.SetMaxOpenConns(dfg.MaxOpenConn)
	sqlDB.SetConnMaxLifetime(dfg.MaxLifeTime)
	sqlDB.SetConnMaxIdleTime(dfg.MaxIdleTime)

	mon := agtsrv.Handler(db, link, slog)
	hub := mlink.Hub(db, link, mon, slog)
	gw := mlink.Gateway(hub)
	listen.BindTo(sh, gw, node)

	mgt := mgtcli.Handler(hub, slog)
	dm := &daemon{
		slog:   slog,
		handle: mgt,
		link:   link,
		ech:    ech,
		ctx:    ctx,
	}

	dm.Run()
}

type daemon struct {
	slog   logback.Logger
	handle http.Handler
	link   telmgt.Linker
	ech    chan<- error
	ctx    context.Context
}

func (d *daemon) Run() {
over:
	for {
		srv := &http.Server{Handler: d.handle}
		lis := d.link.Listen()
		_ = srv.Serve(lis)
		_ = srv.Close()
		d.slog.Warn("新 broker 断开了连接！")

		if err := d.ctx.Err(); err != nil {
			d.ech <- err
			break over
		}

		d.slog.Warn("新 broker 准备重连")
		if err := d.link.Reconnect(d.ctx); err != nil {
			d.ech <- err
			break over
		}
		d.slog.Info("新 broker 重连成功")
	}
}
