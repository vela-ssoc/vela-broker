package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/vela-broker/infra/banner"
	"github.com/vela-ssoc/vela-broker/launch"
)

// args 命令行参数
var args struct {
	version bool   // 打印版本号退出
	config  string // 配置文件 dev 模式才有效
}

func main() {
	parse() // parse 命令行参数
	if banner.Print(os.Stdout); args.version {
		return
	}

	// 初始化日志
	slog := logback.Stdout()
	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()
	slog.Infof("按 Ctrl+C 结束运行")

	if err := launch.Run(ctx, args.config, slog); err != nil {
		slog.Warnf("程序运行错误：%v", err)
	}

	slog.Warnf("程序运行结束")
}
