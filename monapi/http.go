package monapi

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

func Handler() http.Handler {
	sh := ship.Default()
	group := sh.Group("/api")

	new(pingCtrl).BindRoute(group)

	return sh
}
