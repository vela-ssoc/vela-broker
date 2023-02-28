package route

import "github.com/xgfone/ship/v5"

func Ping() RegRouter {
	return &pingCTrl{}
}

type pingCTrl struct{}

func (pc *pingCTrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/ping").GET(pc.Ping)
}

func (pc *pingCTrl) Ping(c *ship.Context) error {
	c.Infof("manager ping broker")
	return nil
}
