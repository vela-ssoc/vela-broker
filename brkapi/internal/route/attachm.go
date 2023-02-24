package route

import (
	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/backend-common/opurl"
	"github.com/vela-ssoc/backend-common/pubrr"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

func AttachMinion(hub mlink.Huber, node string) RegRouter {
	upg := pubrr.Upgrade(node)

	return &attachMinionCtrl{
		hub: hub,
		upg: upg,
	}
}

type attachMinionCtrl struct {
	hub mlink.Huber
	upg websocket.Upgrader
}

func (amc *attachMinionCtrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/mrr/:mid/*path").Any(amc.Forward)
	rgb.Route("/mws/:mid/*path").GET(amc.Socket)
}

func (amc *attachMinionCtrl) Forward(c *ship.Context) error {
	w, r := c.ResponseWriter(), c.Request()
	query := r.URL.RawQuery
	path := c.Param("path")
	mid := c.Param("mid")

	op := opurl.BMrr(mid, c.Method(), path, query)
	amc.hub.Forward(op, w, r)

	return nil
}

func (amc *attachMinionCtrl) Socket(c *ship.Context) error {
	if !c.IsWebSocket() {
		return nil
	}

	w, r := c.ResponseWriter(), c.Request()
	mid := c.Param("mid")
	path := c.Param("path")
	op := opurl.BMws(mid, path, r.URL.RawQuery)
	back, err := amc.hub.Stream(op)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer back.Close()

	fore, err := amc.upg.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer fore.Close()

	pubrr.Pipe(fore, back)

	return nil
}
