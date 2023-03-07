package lisapi

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/validate"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

func Handler(gw http.Handler, hub mlink.Huber, slog logback.Logger) http.Handler {
	node := hub.NodeName()
	sh := ship.Default()
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	sh.Validator = validate.New()
	sh.Logger = slog

	group := sh.Group("/api/v1")
	group.Route("/minion").CONNECT(func(c *ship.Context) error {
		gw.ServeHTTP(c.ResponseWriter(), c.Request())
		return nil
	})

	return sh
}
