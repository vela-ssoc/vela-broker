package route

import (
	"net/http"

	"github.com/xgfone/ship/v5"
)

func Ping() RegRouter {
	return &pingCTrl{}
}

type pingCTrl struct{}

func (pc *pingCTrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/ping").GET(pc.Ping)
}

func (pc *pingCTrl) Ping(c *ship.Context) error {
	c.Infof("manager ping broker")

	w := c.ResponseWriter()
	jack, ok := w.(http.Hijacker)
	c.Infof("hijacker: %v, %v", ok, jack)

	return nil
}
