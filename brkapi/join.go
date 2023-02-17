package brkapi

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

type joinCtrl struct {
	gateway http.Handler
}

func (jc *joinCtrl) BindRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/minion").CONNECT(jc.Join)
}

func (jc *joinCtrl) Join(c *ship.Context) error {
	jc.gateway.ServeHTTP(c.ResponseWriter(), c.Request())
	return nil
}
