package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/vela-ssoc/broker/infra/banner"
	"github.com/vela-ssoc/broker/infra/logback"
	"github.com/vela-ssoc/broker/launch"
)

// args 命令行参数
var args struct {
	version bool   // 打印版本号退出
	config  string // 配置文件 dev 模式才有效
}

func main() {
	parse() // parse 命令行参数
	if banner.Print(); args.version {
		return
	}

	logger := logback.New()
	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()

	if err := launch.Run(ctx, args.config, logger); err != nil {
		logger.Warnf("程序运行错误：%v", err)
	}

	logger.Warnf("程序运行结束")
}
