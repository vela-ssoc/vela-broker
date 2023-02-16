package brkcli

import (
	"net"
	"strings"
	"time"

	"github.com/vela-ssoc/broker/infra/encipher"
	"github.com/vela-ssoc/broker/libkit/credent"
)

// Ident broker 节点握手认证时需要携带的信息
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

// Encrypt 对信息进行加密
func (ident Ident) Encrypt() ([]byte, error) {
	return encipher.EncryptJSON(ident)
}

// Issue 认证成功后返回的必要信息
type Issue struct {
	Passwd []byte `json:"passwd"` // 通信加密密钥
}

func (issue *Issue) Decrypt(data []byte) error {
	return encipher.DecryptJSON(data, issue)
}

// Listen 本地服务监听配置
type Listen struct {
	Addr string `json:"addr"` // 监听地址 :8080 192.168.1.2:8080
	Cert []byte `json:"cert"` // 证书
	Pkey []byte `json:"pkey"` // 私钥
}

// Certifier 获取证书管理器
// 返回值可以为 nil, nil
// 当 error 为 nil 时说明没有错误
// 当 credent.Certifier 为 nil 时说明没有 TLS 证书
func (ln Listen) Certifier() (credent.Certifier, error) {
	if len(ln.Cert) == 0 || len(ln.Pkey) == 0 {
		return nil, nil
	}

	cert, err := credent.Single(ln.Cert, ln.Pkey)
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
	Servers Addresses `json:"servers"`
}

type Address struct {
	TLS  bool   `json:"tls"`
	Addr string `json:"addr"`
	Name string `json:"name"`
}

// String fmt.Stringer
func (a Address) String() string {
	build := new(strings.Builder)
	if a.TLS {
		build.WriteString("tls://")
	} else {
		build.WriteString("tcp://")
	}
	build.WriteString(a.Addr)

	if name := a.Name; name != "" && a.TLS {
		build.WriteString(", servername: ")
		build.WriteString(name)
	}

	return build.String()
}

type Addresses []*Address

func (ads Addresses) Format() {
	for _, ad := range ads {
		addr := ad.Addr
		_, port, err := net.SplitHostPort(addr)
		if err == nil && port != "" {
			continue
		}
		if ad.TLS {
			ad.Addr = addr + ":443"
		} else {
			ad.Addr = addr + ":80"
		}
	}
}
