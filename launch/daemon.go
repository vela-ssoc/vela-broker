package launch

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/vela-ssoc/vela-broker/brkcli"
	"github.com/vela-ssoc/vela-broker/infra/logback"
)

type daemonServer struct {
	listen  brkcli.Listen // 服务监听配置
	handler http.Handler  // handler
	server  *http.Server  // HTTP 服务
	errCh   chan<- error  // 错误输出
}

func (ds *daemonServer) Run() {
	cert, err := ds.listen.Certifier()
	if err != nil {
		ds.errCh <- err
		return
	}

	srv := &http.Server{
		Addr:    ds.listen.Addr,
		Handler: ds.handler,
	}

	if cert != nil {
		cfg := &tls.Config{GetCertificate: cert.Match}
		srv.TLSConfig = cfg
		err = srv.ListenAndServeTLS("", "")
	} else {
		err = srv.ListenAndServe()
	}

	ds.errCh <- err
}

func (ds *daemonServer) Close() error {
	if srv := ds.server; srv != nil {
		return srv.Close()
	}
	return nil
}

type daemonClient struct {
	brk     brkcli.Broker
	handler http.Handler
	server  *http.Server
	errCh   chan<- error
	slog    logback.Logger
	parent  context.Context
}

func (dc *daemonClient) Run() {
	srv := &http.Server{
		Handler: dc.handler,
	}
	dc.server = srv

	for {
		lis := dc.brk.Listener()
		_ = srv.Serve(lis)
		dc.slog.Warn("与中心端的连接已断开")
		if err := dc.parent.Err(); err != nil {
			dc.errCh <- err
			break
		}
		dc.slog.Info("正在准备重试连接中心端")
		if err := dc.brk.Reconnect(dc.parent); err != nil {
			dc.errCh <- err
			break
		}
		dc.slog.Info("重新连接中心端成功")
	}
}

func (dc *daemonClient) Close() error {
	if srv := dc.server; srv != nil {
		return srv.Close()
	}
	return nil
}
