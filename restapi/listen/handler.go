package listen

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/validate"
	"github.com/vela-ssoc/vela-broker/restapi/listen/logic"
	"github.com/xgfone/ship/v5"
)

func Handler(gw http.Handler, node string, slog logback.Logger) http.Handler {
	sh := ship.Default()
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	sh.Validator = validate.New()
	sh.Logger = slog

	v1api := sh.Group("/api/v1")
	logic.Join(gw).Route(v1api)

	return sh
}

func BindTo(sh *ship.Ship, gw http.Handler, node string) {
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	v1api := sh.Group("/api/v1")
	logic.Join(gw).Route(v1api)
}
