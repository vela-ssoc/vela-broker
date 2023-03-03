package brkapi

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/validate"
	"github.com/vela-ssoc/vela-broker/brkapi/internal/route"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

func Handler(hub mlink.Huber, slog logback.Logger) http.Handler {
	node := hub.NodeName()
	sh := ship.Default()
	sh.Logger = slog
	sh.Validator = validate.New()
	sh.HandleError = netutil.ErrorFunc(node)
	sh.NotFound = netutil.Notfound(node)

	group := sh.Group("/api")

	route.Ping().RegRoute(group)
	route.Attach(hub).RegRoute(group)
	route.Pprof().RegRoute(group)

	return sh
}
