package brkcli

import (
	"net"
	"strings"
	"time"

	"github.com/vela-ssoc/broker/libkit/credent"
)

type Issue struct{}

type Ident struct {
	ID         int64     `json:"id"`         // ID
	Secret     string    `json:"secret"`     // 密钥
	IP         net.IP    `json:"ip"`         // IP 地址
	MAC        string    `json:"mac"`        // MAC 地址
	Semver     string    `json:"semver"`     // 版本
	Goos       string    `json:"goos"`       // runtime.GOOS
	Arch       string    `json:"arch"`       // runtime.GOARCH
	CPU        int       `json:"cpu"`        // runtime.NumCPU
	PID        int       `json:"pid"`        // os.Getpid
	Workdir    string    `json:"workdir"`    // os.Getwd
	Executable string    `json:"executable"` // os.Executable
	Username   string    `json:"username"`   // user.Current
	Hostname   string    `json:"hostname"`   // os.Hostname
	TimeAt     time.Time `json:"time_at"`    // 发起时间
}

type Listen struct {
	Addr string `json:"addr"` // 监听地址 :8080 192.168.1.2:8080
	Cert []byte `json:"cert"` // 证书
	Key  []byte `json:"key"`  // 私钥
}

// Certifier 获取证书管理器
// 返回值可以为 nil, nil
// 当 error 为 nil 时说明没有错误
// 当 credent.Certifier 为 nil 时说明没有 TLS 证书
func (ln Listen) Certifier() (credent.Certifier, error) {
	if len(ln.Cert) == 0 || len(ln.Key) == 0 {
		return nil, nil
	}

	cert, err := credent.Single(ln.Cert, ln.Key)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

type Logged struct {
	Console bool `json:"console"`
}

type Database struct {
	DSN string `json:"dsn"`
}

type Config struct {
	Name     string   `json:"name"`
	Listen   Listen   `json:"listen"`   // 监听配置
	Logged   Logged   `json:"logged"`   // 日志配置
	Database Database `json:"database"` // 数据库配置
}

type Hide struct {
	ID      int64     `json:"id"`
	Secret  string    `json:"secret"`
	Semver  string    `json:"semver"`
	Servers []*Server `json:"servers"`
}

type Server struct {
	TLS  bool   `json:"tls"`
	Addr string `json:"addr"`
	Name string `json:"name"`
}

func (srv Server) String() string {
	build := new(strings.Builder)
	if srv.TLS {
		build.WriteString("tls://")
	} else {
		build.WriteString("tcp://")
	}
	build.WriteString(srv.Addr)

	if name := srv.Name; name != "" && srv.TLS {
		build.WriteString(", servername: ")
		build.WriteString(name)
	}

	return build.String()
}
