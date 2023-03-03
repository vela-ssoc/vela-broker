package route

import (
	"net/http/pprof"

	"github.com/xgfone/ship/v5"
)

func Pprof() RegRouter {
	return new(pprofCtrl)
}

type pprofCtrl struct{}

func (pc *pprofCtrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/brr/debug/pprof").GET(pc.Index)
	rgb.Route("/brr/debug/cmdline").GET(pc.Cmdline)
	rgb.Route("/brr/debug/profile").GET(pc.Profile)
	rgb.Route("/brr/debug/symbol").GET(pc.Symbol).POST(pc.Symbol)
	rgb.Route("/brr/debug/trace").GET(pc.Trace)
	rgb.Route("/brr/debug/*name").GET(pc.Lookup)
}

func (pc *pprofCtrl) Index(c *ship.Context) error {
	pprof.Index(c.ResponseWriter(), c.Request())
	return nil
}

func (pc *pprofCtrl) Cmdline(c *ship.Context) error {
	pprof.Cmdline(c.ResponseWriter(), c.Request())
	return nil
}

func (pc *pprofCtrl) Profile(c *ship.Context) error {
	pprof.Profile(c.ResponseWriter(), c.Request())
	return nil
}

func (pc *pprofCtrl) Lookup(c *ship.Context) error {
	name := c.Param("name")
	pprof.Handler(name).ServeHTTP(c.ResponseWriter(), c.Request())
	return nil
}

func (pc *pprofCtrl) Symbol(c *ship.Context) error {
	pprof.Symbol(c.ResponseWriter(), c.Request())
	return nil
}

func (pc *pprofCtrl) Trace(c *ship.Context) error {
	pprof.Trace(c.ResponseWriter(), c.Request())
	return nil
}
