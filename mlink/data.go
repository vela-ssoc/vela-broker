package mlink

import (
	"bytes"
	"io"
	"net"
	"time"

	"github.com/vela-ssoc/backend-common/encipher"
)

// Ident minion 节点握手认证时需要携带的信息
type Ident struct {
	Semver     string    `json:"semver"`     // 节点版本
	Inet       net.IP    `json:"inet"`       // 内网出口 IP
	MAC        string    `json:"mac"`        // 出口 IP 所在网卡的 MAC 地址
	Goos       string    `json:"goos"`       // 操作系统 runtime.GOOS
	Arch       string    `json:"arch"`       // 操作系统架构 runtime.GOARCH
	CPU        int       `json:"cpu"`        // CPU 核心数
	PID        int       `json:"pid"`        // 进程 PID
	Workdir    string    `json:"workdir"`    // 工作目录
	Executable string    `json:"executable"` // 执行路径
	Username   string    `json:"username"`   // 当前操作系统用户名
	Hostname   string    `json:"hostname"`   // 主机名
	TimeAt     time.Time `json:"time_at"`    // 当前时间，暂无太大意义
}

// Decrypt 认证身份信息解密
func (ide *Ident) Decrypt(enc []byte) error {
	return encipher.DecryptJSON(enc, ide)
}

// Issue 信息
type Issue struct {
	ID     int64  `json:"id"`
	Passwd []byte `json:"passwd"`
}

func (iss Issue) Encrypt() ([]byte, error) {
	return encipher.EncryptJSON(iss)
}

type jsonBody struct {
	err error
	buf *bytes.Buffer
}

func (jb *jsonBody) Read(p []byte) (int, error) {
	if jb.err != nil {
		return 0, jb.err
	}
	if jb.buf == nil {
		return 0, io.EOF
	}
	return jb.buf.Read(p)
}

func (jb *jsonBody) Close() error { return nil }

func (jb *jsonBody) Len() int {
	if jb.err != nil || jb.buf == nil {
		return 0
	}
	return jb.buf.Len()
}
