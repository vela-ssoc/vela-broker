package monapi

import (
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

type pingCtrl struct{}

func (pc *pingCtrl) BindRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/ping").GET(pc.Ping)
}

func (pc *pingCtrl) Ping(c *ship.Context) error {
	infer := mlink.UnwarpCtx(c.Request().Context())
	inet := infer.Ident().IP.String()
	c.Infof("minion %s 发来了心跳包", inet)
	return nil
}
