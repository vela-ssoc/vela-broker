package logic

import (
	"net/http"

	"github.com/vela-ssoc/vela-broker/restapi/facade"
	"github.com/xgfone/ship/v5"
)

func Join(gw http.Handler) facade.Router {
	return &joinCtrl{gw: gw}
}

type joinCtrl struct {
	gw http.Handler
}

func (jc *joinCtrl) Route(r *ship.RouteGroupBuilder) {
	r.Route("/minion").CONNECT(jc.Join)
}

func (jc *joinCtrl) Join(c *ship.Context) error {
	w, r := c.ResponseWriter(), c.Request()
	jc.gw.ServeHTTP(w, r)
	return nil
}
