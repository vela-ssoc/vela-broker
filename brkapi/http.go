package brkapi

import (
	"net/http"

	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

func Handler(hub mlink.Huber) http.Handler {
	sh := ship.Default()
	group := sh.Group("/api")

	group.Route("/ping").GET(func(c *ship.Context) error {
		return c.Text(http.StatusOK, "server PONG")
	})

	return sh
}
