package monapi

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/validate"
	"github.com/vela-ssoc/vela-broker/dialmgt"
	"github.com/vela-ssoc/vela-broker/monapi/internal/route"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
)

func Handler(db *gorm.DB, link dialmgt.Linker, slog logback.Logger) http.Handler {
	node := link.NodeName()
	sh := ship.Default()
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	sh.Validator = validate.New()
	sh.Logger = slog

	group := sh.Group("/api/v1")
	route.Ping(db, link).RegRoute(group)

	return sh
}
