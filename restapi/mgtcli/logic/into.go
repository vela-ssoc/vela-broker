package logic

import (
	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/opurl"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/vela-ssoc/vela-broker/restapi/facade"
	"github.com/xgfone/ship/v5"
)

func Into(hub mlink.Huber) facade.Router {
	upg := netutil.Upgrade(hub.NodeName())
	return &intoLogic{
		hub: hub,
		upg: upg,
	}
}

type intoLogic struct {
	hub mlink.Huber
	upg websocket.Upgrader
}

func (in *intoLogic) Route(r *ship.RouteGroupBuilder) {
	r.Route("/arr/:mid/*path").Any(in.Arr)
	r.Route("/aws/:mid/*path").GET(in.Aws)
}

// Arr agent request response 向 agent 发送请求响应
func (in *intoLogic) Arr(c *ship.Context) error {
	mid := c.Param("mid")   // minion ID
	path := c.Param("path") // 请求路径
	w, r := c.ResponseWriter(), c.Request()
	op := opurl.BArr(mid, c.Method(), path, r.URL.RawQuery)
	in.hub.Forward(op, w, r)
	return nil
}

// Aws agent websocket 转发
func (in *intoLogic) Aws(c *ship.Context) error {
	mid := c.Param("mid")
	path := c.Param("path")
	w, r := c.ResponseWriter(), c.Request()
	query := r.URL.RawQuery
	op := opurl.BAws(mid, path, query)
	back, err := in.hub.Stream(op, nil)
	if err != nil {
		return err
	}

	fore, err := in.upg.Upgrade(w, r, nil)
	if err != nil {
		_ = back.Close()
		return err
	}

	netutil.Pipe(fore, back)

	return nil
}
