package dialmgt

import (
	"net"
	"strings"
)

type Address struct {
	TLS  bool   `json:"tls"  yaml:"tls"`
	Addr string `json:"addr" yaml:"addr"`
	Name string `json:"name" yaml:"name"`
}

// String fmt.Stringer
func (ad Address) String() string {
	build := new(strings.Builder)
	if ad.TLS {
		build.WriteString("tls://")
	} else {
		build.WriteString("tcp://")
	}
	build.WriteString(ad.Addr)

	if name := ad.Name; name != "" && ad.TLS {
		build.WriteString(", servername: ")
		build.WriteString(name)
	}

	return build.String()
}

type Addresses []*Address

// Format 对地址进行格式化处理，即：如果地址内有显式端口号，
// 则根据是否开启 TLS 补充默认端口号
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
