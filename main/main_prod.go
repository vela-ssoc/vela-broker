//go:build !dev

package main

import (
	"flag"
	"os"

	"github.com/vela-ssoc/backend-common/encipher"
	"github.com/vela-ssoc/vela-broker/telmgt"
)

// args 命令行参数
var args struct {
	version bool // 打印版本号退出
}

func parse() {
	flag.BoolVar(&args.version, "v", false, "打印版本号")
	flag.Parse()
}

// readHide 生产环境
func readHide() (telmgt.Hide, error) {
	var hide telmgt.Hide
	err := encipher.ReadFile(os.Args[0], &hide)
	return hide, err
}
