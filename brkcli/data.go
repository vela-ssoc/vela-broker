package brkcli

import (
	"net"
	"strings"
	"time"

	"github.com/vela-ssoc/backend-common/encipher"
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

type Hide struct {
	ID      int64     `json:"id"`      // ID
	Secret  string    `json:"secret"`  // 认证密钥
	Semver  string    `json:"semver"`  // broker 版本
	Servers Addresses `json:"servers"` // 地址
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
