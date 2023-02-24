package brkapi

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/pubrr"
	"github.com/vela-ssoc/vela-broker/brkapi/internal/route"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

func Handler(hub mlink.Huber, slog logback.Logger) http.Handler {
	sh := ship.Default()
	sh.Logger = slog

	inet := hub.BrkInet()
	node := "broker-" + inet.String()
	sh.HandleError = pubrr.ErrorHandle(node)
	sh.NotFound = pubrr.NotFound(node)

	group := sh.Group("/api")
	route.Ping().RegRoute(group)
	route.Attach().RegRoute(group)
	route.AttachMinion(hub, node).RegRoute(group)

	return sh
}
