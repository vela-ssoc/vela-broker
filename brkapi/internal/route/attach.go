package route

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/opurl"
	"github.com/vela-ssoc/backend-common/pubody"
	"github.com/vela-ssoc/backend-common/syscmd"
	"github.com/vela-ssoc/vela-broker/mlink"
	"github.com/xgfone/ship/v5"
)

func Attach(hub mlink.Huber) RegRouter {
	node := hub.NodeName()
	upg := netutil.Upgrade(node)

	return &attachCtrl{
		hub: hub,
		upg: upg,
	}
}

type attachCtrl struct {
	hub mlink.Huber
	upg websocket.Upgrader
}

func (sc *attachCtrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/arr/:mid/*path").Any(sc.Arr)
	rgb.Route("/aws/:mid/*path").GET(sc.Aws)
	rgb.Route("/brr/syscmd").GET(sc.Syscmd)
	rgb.Route("/brr/fm").GET(sc.FM)
	rgb.Route("/bws/echo").GET(sc.Echo)
}

func (sc *attachCtrl) Arr(c *ship.Context) error {
	mid := c.Param("mid")
	path := c.Param("path")
	w, r := c.ResponseWriter(), c.Request()
	op := opurl.BArr(mid, c.Method(), path, r.URL.RawQuery)
	sc.hub.Forward(op, w, r)

	return nil
}

func (sc *attachCtrl) Aws(c *ship.Context) error {
	mid := c.Param("mid")
	path := c.Param("path")
	w, r := c.ResponseWriter(), c.Request()
	query := r.URL.RawQuery
	op := opurl.BAws(mid, path, query)
	back, err := sc.hub.Stream(op, nil)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer back.Close()

	fore, err := sc.upg.Upgrade(w, r, nil)
	if err != nil {
		return err
	}

	netutil.Pipe(fore, back)

	return nil
}

func (sc *attachCtrl) Syscmd(c *ship.Context) error {
	var req syscmd.Input
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	ret := syscmd.Exec(req, false)
	c.Warnf("[命令执行] %s %s, error: %s", ret.Cmd, strings.Join(ret.Args, " "), ret.Error)

	return c.JSON(http.StatusOK, ret)
}

func (sc *attachCtrl) FM(c *ship.Context) error {
	match := c.Query("match")
	str := c.Query("path", "./")
	abs, err := filepath.Abs(str)
	if err != nil {
		return err
	}
	open, err := os.Open(abs)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer open.Close()

	if stat, err := open.Stat(); err != nil {
		return err
	} else if !stat.IsDir() {
		return c.Attachment(abs, "")
	}

	infos, err := open.Readdir(-1)
	if err != nil {
		return err
	}

	sep := string(os.PathSeparator)
	ret := &pubody.Folder{
		Abs:       abs,
		Separator: sep,
		Items:     make(pubody.FileItems, 0, 32),
	}
	var matchErr error
	for _, info := range infos {
		nm := info.Name()
		if matchErr == nil && match != "" {
			var matched bool
			matched, matchErr = filepath.Match(match, nm)
			if matchErr == nil && !matched {
				continue
			}
		}
		dir := info.IsDir()
		item := &pubody.FileItem{
			Path:  filepath.Join(abs, nm),
			Name:  nm,
			Size:  info.Size(),
			Mtime: info.ModTime(),
			Dir:   dir,
			Mode:  info.Mode().String(),
		}
		if !dir {
			item.Ext = filepath.Ext(nm)
		}
		ret.Items = append(ret.Items, item)
	}

	return c.JSON(http.StatusOK, ret)
}

func (sc *attachCtrl) Echo(c *ship.Context) error {
	w, r := c.ResponseWriter(), c.Request()
	conn, err := sc.upg.Upgrade(w, r, nil)
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer conn.Close()

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			c.Warnf("读取 socket 错误：%s", err)
			break
		}
		str := string(data)
		c.Infof("socket 收到消息：%s", str)
		ret := &socketReply{Type: mt, Data: str, TimeAt: time.Now()}
		if err = conn.WriteJSON(ret); err != nil {
			c.Warnf("写入 socket 错误：%s", err)
			break
		}
	}

	return nil
}

type socketReply struct {
	Type   int       `json:"type"`
	Data   string    `json:"data"`
	TimeAt time.Time `json:"time_at"`
}
