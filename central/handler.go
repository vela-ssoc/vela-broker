package central

import (
	"net/http"
	"os"

	"github.com/xgfone/ship/v5"
	"golang.org/x/net/webdav"
)

func New() http.Handler {
	sh := ship.Default()
	group := sh.Group("/api")
	var emp empty
	group.Route("/env").GET(emp.Env)
	group.Route("/")

	// WebDAV
	methods := []string{
		http.MethodOptions, http.MethodGet, http.MethodHead, http.MethodPost, http.MethodDelete,
		http.MethodPut, "MKCOL", "COPY", "MOVE", "LOCK", "UNLOCK", "PROPFIND", "PROPPATCH",
	}
	dfs := &webdav.Handler{
		FileSystem: webdav.Dir("/"),
		LockSystem: webdav.NewMemLS(),
	}
	dav := &davFS{h: dfs}
	group.Route("/webdav/broker").Method(dav.Broker, methods...)
	group.Route("/webdav/broker/*path").Method(dav.Broker, methods...)

	return sh
}

type empty struct{}

func (empty) Env(c *ship.Context) error {
	envs := os.Environ()
	return c.JSON(http.StatusOK, envs)
}

type davFS struct {
	h http.Handler
}

func (df davFS) Broker(c *ship.Context) error {
	path := c.Param("path")
	req := c.Request()
	req.URL.Path = "/" + path
	w := c.Response()

	df.h.ServeHTTP(w, req)

	return nil
}
