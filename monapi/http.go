package monapi

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/vela-broker/dialmgt"
	"github.com/vela-ssoc/vela-broker/monapi/internal/route"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
)

func Handler(db *gorm.DB, link dialmgt.Linker, slog logback.Logger) http.Handler {
	sh := ship.Default()
	sh.Logger = slog

	group := sh.Group("/api")
	route.Ping(db, link).RegRoute(group)

	return sh
}
