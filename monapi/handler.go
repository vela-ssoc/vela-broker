package monapi

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

func NewHandler(gateway http.Handler) http.Handler {
	sh := ship.Default()
	group := sh.Group("/api")

	newJoin(gateway).BindRoute(group)

	return sh
}
