package listen

import (
	"net/http"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/validate"
	"github.com/xgfone/ship/v5"
)

func Handler(gw http.Handler, node string, slog logback.Logger) http.Handler {
	sh := ship.Default()
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	sh.Validator = validate.New()
	sh.Logger = slog

	v1api := sh.Group("/api/v1")

	mj := &minionJoin{gw: gw}
	v1api.Route("/minion").CONNECT(mj.Join)

	return sh
}

// BindTo 过度接口
//
// Deprecated: 临时兼容启动，以后会下线该版本
func BindTo(sh *ship.Ship, gw http.Handler, node string) {
	sh.NotFound = netutil.Notfound(node)
	sh.HandleError = netutil.ErrorFunc(node)
	v1api := sh.Group("/api/v1")

	mj := &minionJoin{gw: gw}
	v1api.Route("/minion").CONNECT(mj.Join)
}

type minionJoin struct {
	gw http.Handler
}

func (mj *minionJoin) Join(c *ship.Context) error {
	w, r := c.ResponseWriter(), c.Request()
	mj.gw.ServeHTTP(w, r)
	return nil
}
