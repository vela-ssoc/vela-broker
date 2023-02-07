package banner

import (
	"fmt"
	"os"
	"runtime"
)

const logo = "\u001B[1;33m" +
	"   ______________  _____\n" +
	"  / ___/ ___/ __ \\/ ___/\n" +
	" (__  |__  ) /_/ / /__  \n" +
	"/____/____/\\____/\\___/  \u001B[0m  \u001B[1;32mBROKER\u001B[0m\n" +
	"Powered By: 东方财富安全团队\n\n" +
	"\t进程 PID: %d\n" +
	"\t操作系统: %s\n" +
	"\t系统架构: %s\n\n"

func Print() {
	fmt.Printf(logo, os.Getpid(), runtime.GOOS, runtime.GOARCH)
}
