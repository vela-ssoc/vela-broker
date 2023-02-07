package brkcli

import "net/http"

var (
	// brkJoin broker 节点认证接入接口
	brkJoin = &opera{method: http.MethodConnect, path: "/api/blink", desc: "认证握手"}
	BrkPing = &opera{method: http.MethodGet, path: "/api/ping", desc: "PING 中心端"}
	BrkStat = &opera{method: http.MethodPost, path: "/api/stat", desc: "信息采集上报"}
)

type Operator interface {
	Method() string
	Path() string
	Desc() string
}

type opera struct {
	method, path, desc string
}

func (op *opera) Method() string { return op.method }
func (op *opera) Path() string   { return op.path }
func (op *opera) Desc() string   { return op.desc }
