package route

import (
	"github.com/vela-ssoc/backend-common/opurl"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

func Intom(hub mlink.Huber) RegRouter {
	return &intomCtrl{hub: hub}
}

type intomCtrl struct {
	hub mlink.Huber
}

func (ic *intomCtrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/intom/:mid/*path").Any(ic.Into)
}

func (ic *intomCtrl) Warp(fn func(*ship.Context, string) error) ship.Handler {
	return func(c *ship.Context) error {
		mid := c.Param("mid")
		return fn(c, mid)
	}
}

func (ic *intomCtrl) Into(c *ship.Context) error {
	w, r := c.ResponseWriter(), c.Request()
	query := r.URL.RawQuery
	path := c.Param("path")
	mid := c.Param("mid")

	op := opurl.BIntom(mid, c.Method(), path, query)
	ic.hub.Forward(op, w, r)

	return nil
}
