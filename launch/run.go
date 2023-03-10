package launch

import (
	"context"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/vela-ssoc/vela-broker/restapi/agtsrv"
	"github.com/vela-ssoc/vela-broker/restapi/listen"
	"github.com/vela-ssoc/vela-broker/restapi/mgtcli"
	"github.com/vela-ssoc/vela-broker/telmgt"
	"go.uber.org/zap"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Run 运行服务
func Run(parent context.Context, hide telmgt.Hide, slog logback.Logger) error {
	link, err := telmgt.Dial(parent, hide, slog) // 与中心端建立连接
	if err != nil {
		return err
	}

	ident := link.Ident()
	issue := link.Issue()
	slog.Infof("broker 接入认证成功，上报认证信息如下：\n%s\n下发的配置如下：\n%s", ident, issue)

	zlg := issue.Logger.Zap()                           // 根据配置文件初始化日志
	slog.Replace(zlg.WithOptions(zap.AddCallerSkip(1))) // 替换日志输出内核

	dbCfg := issue.Database
	glg := logback.Gorm(zlg, dbCfg.Level)
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

	mon := agtsrv.Handler(db, link, slog)
	hub := mlink.Hub(db, link, mon, slog)
	gateway := mlink.Gateway(hub)
	_ = hub.ResetDB()

	errCh := make(chan error, 1)

	// 监听本地端口用于 minion 节点连接
	node := hub.NodeName()
	lis := listen.Handler(gateway, node, slog)
	ds := &daemonServer{listen: issue.Listen, handler: lis, errCh: errCh}
	go ds.Run()

	// 连接 manager 的客户端，保持在线与接受指令
	suborder := mgtcli.Handler(hub, slog)
	dc := &daemonClient{link: link, handler: suborder, errCh: errCh, slog: slog, parent: parent}
	go dc.Run()

	select {
	case err = <-errCh:
	case <-parent.Done():
	}

	_ = ds.Close()
	_ = dc.Close()
	_ = hub.ResetDB()
	_ = rawDB.Close()
	_ = zlg.Sync()

	return err
}
