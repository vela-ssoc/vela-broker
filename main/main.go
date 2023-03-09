package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/vela-broker/banner"
	"github.com/vela-ssoc/vela-broker/launch"
)

func main() {
	parse() // parse 命令行参数
	if banner.Print(os.Stdout); args.version {
		return
	}

	slog := logback.Stdout() // 初始化日志
	hide, err := readHide()  // 读取 hide 配置
	if err != nil {
		slog.Warnf("读取 hide 配置错误：%v", err)
	}

	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()
	slog.Infof("按 Ctrl+C 结束运行")

	if err = launch.Run(ctx, hide, slog); err != nil {
		slog.Warnf("程序运行错误：%v", err)
	}

	slog.Info("程序运行结束")
}
