package brkapi

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

func Handler() http.Handler {
	sh := ship.Default()
	group := sh.Group("/api")

	group.Route("/ping").GET(func(c *ship.Context) error {
		return c.Text(http.StatusOK, "server PONG")
	})

	return sh
}
