package launch

// CRun 与老的 broker 兼容性接口，升级后会删除。
//
// Deprecated: use Run
//func CRun(ctx context.Context, hide dialmgt.Hide, sh *ship.Ship, slog logback.Logger) error {
//	link, err := dialmgt.Dial(ctx, hide, slog)
//	if err != nil {
//		return err
//	}
//	ident, issue := link.Ident(), link.Issue()
//	slog.Infof("broker 接入认证成功，上报认证信息如下：\n%s\n下发的配置如下：\n%s", ident, issue)
//
//	zlg := issue.Logger.Zap()
//	slog.Replace(zlg.WithOptions(zap.AddCallerSkip(1)))
//
//	dfg := issue.Database
//	lvl, dsn := dfg.Level, dfg.FormatDSN()
//	dlg := logback.Gorm(zlg, lvl)
//	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: dlg})
//	if err != nil {
//		return err
//	}
//	sqlDB, err := db.DB()
//	if err != nil {
//		return err
//	}
//	sqlDB.SetMaxIdleConns(dfg.MaxIdleConn)
//	sqlDB.SetMaxOpenConns(dfg.MaxOpenConn)
//	sqlDB.SetConnMaxLifetime(dfg.MaxLifeTime)
//	sqlDB.SetConnMaxIdleTime(dfg.MaxIdleTime)
//
//	mon := monapi.Handler(db, link, slog)
//	hub := mlink.Hub(db, link, mon, slog)
//	gw := mlink.Gateway(hub)
//	// 监听本地端口用于 minion 节点连接
//	local := lisapi.Handler(gw, hub, slog)
//
//	return nil
//}
