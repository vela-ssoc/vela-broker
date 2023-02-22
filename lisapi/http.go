package lisapi

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

func Handler(gw http.Handler) http.Handler {
	sh := ship.Default()
	group := sh.Group("/api")
	group.Route("/minion").CONNECT(func(c *ship.Context) error {
		gw.ServeHTTP(c.ResponseWriter(), c.Request())
		return nil
	})

	return sh
}
