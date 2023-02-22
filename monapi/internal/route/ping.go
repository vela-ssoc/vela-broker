package route

import (
	"time"

	"github.com/vela-ssoc/backend-common/model"
	"github.com/vela-ssoc/vela-broker/dialmgt"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
)

func Ping(db *gorm.DB, link dialmgt.Linker) RegRouter {
	return &pingCtrl{
		db:   db,
		link: link,
	}
}

type pingCtrl struct {
	db   *gorm.DB
	link dialmgt.Linker
}

func (pc *pingCtrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/ping").GET(pc.Ping)
}

func (pc *pingCtrl) Ping(c *ship.Context) error {
	infer := mlink.Ctx(c.Request().Context())
	// 获取 minion 节点 ID
	issue := infer.Issue()
	mid := issue.ID
	ident := infer.Ident()
	inet := ident.Inet
	bid := pc.link.Ident().ID

	c.Infof("minion %s(%d) 发来了心跳包", inet, mid)

	return pc.db.Model(&model.Minion{ID: mid}).
		Where("status = ? AND broker_id = ?", model.MinionOnline, bid).
		UpdateColumn("pinged_at", time.Now()).
		Error
}
