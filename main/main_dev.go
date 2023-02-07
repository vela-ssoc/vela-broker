//go:build dev

package main

import "flag"

func parse() {
	flag.BoolVar(&args.version, "v", false, "打印版本号")
	flag.StringVar(&args.config, "c", "broker.json", "加载配置文件")
	flag.Parse()
}
