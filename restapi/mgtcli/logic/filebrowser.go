package logic

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/vela-ssoc/backend-common/pubody"
	"github.com/vela-ssoc/vela-broker/restapi/facade"
	"github.com/xgfone/ship/v5"
)

func FileBrowser() facade.Router {
	return new(fileBrowserLogic)
}

type fileBrowserLogic struct{}

func (fb *fileBrowserLogic) Route(r *ship.RouteGroupBuilder) {
	r.Route("/brr/fm").
		HEAD(fb.Head).
		GET(fb.Browser)
}

func (fb *fileBrowserLogic) Head(c *ship.Context) error {
	path := c.Query("path")
	if path == "" {
		if path, _ = os.UserHomeDir(); path == "" {
			path = os.TempDir()
		}
	}

	w, r := c.ResponseWriter(), c.Request()
	http.ServeFile(w, r, path)

	return nil
}

func (fb *fileBrowserLogic) Browser(c *ship.Context) error {
	path := c.Query("path")
	lim := c.Query("limit")
	match := c.Query("match")
	if path == "" {
		if path, _ = os.UserHomeDir(); path == "" {
			path = os.TempDir()
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	open, err := os.Open(abs)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer open.Close()
	stat, err := open.Stat()
	if err != nil {
		return err
	}
	if !stat.IsDir() {
		w, r := c.ResponseWriter(), c.Request()
		http.ServeContent(w, r, open.Name(), stat.ModTime(), open)
		return nil
	}

	limit, _ := strconv.ParseInt(lim, 10, 64)
	infos, err := open.Readdir(int(limit))
	if err != nil {
		return err
	}
	sep := string(os.PathSeparator)
	ret := &pubody.Folder{
		Abs:       abs,
		Separator: sep,
		Items:     make(pubody.FileItems, 0, 32),
	}
	var mee error
	var matched bool
	for _, info := range infos {
		name := info.Name()
		if mee == nil && match != "" {
			matched, mee = filepath.Match(match, name)
			if mee == nil && !matched {
				continue
			}
		}
		dir := info.IsDir()
		item := &pubody.FileItem{
			Path:  filepath.Join(abs, name),
			Name:  name,
			Size:  info.Size(),
			Mtime: info.ModTime(),
			Dir:   dir,
			Mode:  info.Mode().String(),
		}
		if !dir {
			item.Ext = filepath.Ext(name)
		}
		ret.Items = append(ret.Items, item)
	}

	return c.JSON(http.StatusOK, ret)
}
