package brkapi

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

func NewHandler() http.Handler {
	sh := ship.Default()
	sh.Route("/ping").GET(func(c *ship.Context) error {
		return c.Text(http.StatusOK, "server PONG")
	})

	return sh
}
