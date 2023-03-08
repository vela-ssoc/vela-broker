package mgtcli

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/validate"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/vela-ssoc/vela-broker/restapi/mgtcli/logic"
	"github.com/xgfone/ship/v5"
)

func Handler(hub mlink.Huber, slog logback.Logger) http.Handler {
	node := hub.NodeName()
	sh := ship.Default()
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	sh.Validator = validate.New()
	sh.Logger = slog

	v1api := sh.Group("/api/v1")
	logic.Pprof().Route(v1api)
	logic.Into(hub).Route(v1api)
	logic.FileBrowser().Route(v1api)

	return sh
}
