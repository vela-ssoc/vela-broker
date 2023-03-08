package logic

import (
	"net/http/pprof"

	"github.com/vela-ssoc/vela-broker/restapi/facade"
	"github.com/xgfone/ship/v5"
)

func Pprof() facade.Router {
	return new(pprofLogic)
}

type pprofLogic struct{}

func (pf *pprofLogic) Route(r *ship.RouteGroupBuilder) {
	r.Route("/brr/debug/pprof").GET(pf.Index)
	r.Route("/brr/debug/cmdline").GET(pf.Cmdline)
	r.Route("/brr/debug/profile").GET(pf.Profile)
	r.Route("/brr/debug/symbol").GET(pf.Symbol).POST(pf.Symbol)
	r.Route("/brr/debug/trace").GET(pf.Trace)
	r.Route("/brr/debug/*name").GET(pf.Lookup)
}

func (pf *pprofLogic) Index(c *ship.Context) error {
	pprof.Index(c.ResponseWriter(), c.Request())
	return nil
}

func (pf *pprofLogic) Cmdline(c *ship.Context) error {
	pprof.Cmdline(c.ResponseWriter(), c.Request())
	return nil
}

func (pf *pprofLogic) Profile(c *ship.Context) error {
	pprof.Profile(c.ResponseWriter(), c.Request())
	return nil
}

func (pf *pprofLogic) Lookup(c *ship.Context) error {
	name := c.Param("name")
	pprof.Handler(name).ServeHTTP(c.ResponseWriter(), c.Request())
	return nil
}

func (pf *pprofLogic) Symbol(c *ship.Context) error {
	pprof.Symbol(c.ResponseWriter(), c.Request())
	return nil
}

func (pf *pprofLogic) Trace(c *ship.Context) error {
	pprof.Trace(c.ResponseWriter(), c.Request())
	return nil
}
