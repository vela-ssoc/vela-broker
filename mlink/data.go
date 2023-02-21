package mlink

import (
	"net"
	"time"

	"github.com/vela-ssoc/backend-common/encipher"
)

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

func (ide *Ident) Decrypt(enc []byte) error {
	return encipher.DecryptJSON(enc, ide)
}

type Issue struct {
	ID     int64  `json:"id"`
	Passwd []byte `json:"passwd"`
}

func (iss Issue) Encrypt() ([]byte, error) {
	return encipher.EncryptJSON(iss)
}
