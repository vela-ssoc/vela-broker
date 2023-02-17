package launch

import (
	"context"
	"os"

	"github.com/vela-ssoc/broker/brkcli"
	"github.com/vela-ssoc/broker/guard"
	"github.com/vela-ssoc/broker/infra/bootstrap"
	"github.com/vela-ssoc/broker/infra/logback"
	"github.com/vela-ssoc/broker/minister"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Run 运行服务
func Run(parent context.Context, cfg string, slog logback.Logger) error {
	var hide brkcli.Hide
	if err := bootstrap.AutoLoad(cfg, os.Args[0], &hide); err != nil {
		return err
	}

	brk, err := brkcli.MustJoin(parent, hide, slog)
	if err != nil {
		return err
	}
	issue := brk.Issue()
	slog.Infof("broker 接入认证成功，下发的配置如下：\n%s", issue)

	zlg := issue.Logger.Zap()                           // 根据配置文件初始化日志
	slog.Replace(zlg.WithOptions(zap.AddCallerSkip(1))) // 替换日志输出内核

	dbCfg := issue.Database
	glg := logback.GORM(zlg, dbCfg.Level)
	db, err := gorm.Open(mysql.Open(dbCfg.FormatDSN()), &gorm.Config{Logger: glg})
	if err != nil {
		return err
	}
	rawDB, err := db.DB()
	if err != nil {
		return err
	}
	rawDB.SetMaxIdleConns(dbCfg.MaxIdleConn)
	rawDB.SetMaxOpenConns(dbCfg.MaxOpenConn)
	rawDB.SetConnMaxLifetime(dbCfg.MaxLifeTime)
	rawDB.SetConnMaxIdleTime(dbCfg.MaxIdleTime)

	serve := minister.NewHandler()
	lisCfg := issue.Listen
	errCh := make(chan error, 1)
	ds := &daemonServer{listen: lisCfg, handler: serve, errCh: errCh}
	go ds.Run()

	suborder := guard.NewHandler()
	dc := &daemonClient{brk: brk, handler: suborder, errCh: errCh, slog: slog, parent: parent}
	go dc.Run()

	select {
	case err = <-errCh:
	case <-parent.Done():
		err = parent.Err()
	}

	_ = ds.Close()
	_ = dc.Close()
	_ = rawDB.Close()
	_ = zlg.Sync()

	return err
}
