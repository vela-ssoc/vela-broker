package mgtapi

import (
	"github.com/vela-ssoc/vela-broker/app/route"
	"github.com/vela-ssoc/vela-common-mb/accord"
	"github.com/vela-ssoc/vela-common-mb/integration/cmdb"
	"github.com/vela-ssoc/vela-common-mb/integration/elastic"
	"github.com/xgfone/ship/v5"
)

func Reset(esCfg elastic.Configurer, cmdbCfg cmdb.Configurer) route.Router {
	return &resetREST{
		esCfg:   esCfg,
		cmdbCfg: cmdbCfg,
	}
}

type resetREST struct {
	esCfg   elastic.Configurer
	cmdbCfg cmdb.Configurer
}

func (rest *resetREST) Route(r *ship.RouteGroupBuilder) {
	r.Route(accord.PathElasticReset).Data(route.Named("elastic 配置 reset")).POST(rest.Elastic)
	r.Route(accord.PathCmdbReset).Data(route.Named("cmdb 配置 reset")).POST(rest.Cmdb)
}

func (rest *resetREST) Elastic(c *ship.Context) error {
	rest.esCfg.Reset()
	c.Infof("elastic 配置 reset")
	return nil
}

func (rest *resetREST) Cmdb(c *ship.Context) error {
	rest.cmdbCfg.Reset()
	c.Infof("cmdb 配置 reset")
	return nil
}
