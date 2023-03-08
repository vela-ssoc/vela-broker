package agtsrv

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/validate"
	"github.com/vela-ssoc/vela-broker/telmgt"
	"github.com/xgfone/ship/v5"
	"gorm.io/gorm"
)

func Handler(db *gorm.DB, link telmgt.Linker, slog logback.Logger) http.Handler {
	node := link.NodeName()
	sh := ship.Default()
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	sh.Validator = validate.New()
	sh.Logger = slog

	group := sh.Group("/api/v1")
	_ = group

	return sh
}
