//go:build dev

package main

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/vela-ssoc/vela-broker/telmgt"
)

// args 命令行参数
var args struct {
	version bool   // 打印版本号退出
	config  string // 配置文件 dev 模式才有效
}

func parse() {
	flag.BoolVar(&args.version, "v", false, "打印版本号")
	flag.StringVar(&args.config, "c", "broker.json", "加载配置文件")
	flag.Parse()
}

func readHide() (telmgt.Hide, error) {
	var hide telmgt.Hide
	open, err := os.Open(args.config)
	if err != nil {
		return hide, err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer open.Close()

	err = json.NewDecoder(open).Decode(&hide)

	return hide, err
}
