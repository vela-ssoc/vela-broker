package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/vela-ssoc/vela-broker/banner"
	"github.com/vela-ssoc/vela-broker/bridge/telecom"
	"github.com/vela-ssoc/vela-broker/launch"
	"github.com/vela-ssoc/vela-common-mb/logback"
	"github.com/vela-ssoc/vela-common-mb/ulimit"
	"github.com/vela-ssoc/vela-common-mba/ciphertext"
)

var args struct {
	// dev 是否开发环境，开发环境时请在编译时指定 go build -tags=dev
	dev     bool
	version bool
	config  string
}

func main() {
	flag.BoolVar(&args.version, "v", false, "打印版本号")
	if args.dev {
		// 开启开发环境，允许自定义连接配置
		flag.StringVar(&args.config, "c", "broker.json", "加载配置文件")
	}
	flag.Parse()
	if banner.WriteTo(os.Stdout); args.version {
		return
	}

	slog := logback.Stdout() // 初始化日志
	hide, err := loadHide()  // 读取 hide 配置
	if err != nil {
		slog.Warnf("读取 hide 配置错误：%v", err)
		return
	}

	cares := []os.Signal{syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL, syscall.SIGINT}
	ctx, cancel := signal.NotifyContext(context.Background(), cares...)
	defer cancel()
	slog.Infof("按 Ctrl+C 结束运行")

	const limit = 65536
	if err = ulimit.Least(limit); err != nil {
		slog.Warnf("调整文件描述符个数为 %d 时错误: %v", limit, err)
	}

	if err = launch.Run(ctx, hide, slog); err != nil {
		slog.Warnf("程序运行错误：%v", err)
	}

	slog.Info("程序运行结束")
}

func loadHide() (telecom.Hide, error) {
	var hide telecom.Hide
	if !args.dev {
		arg := os.Args[0]
		err := ciphertext.DecryptFile(arg, &hide)
		return hide, err
	}

	file, err := os.Open(args.config)
	if err != nil {
		return hide, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	err = json.NewDecoder(file).Decode(&hide)

	return hide, err
}
