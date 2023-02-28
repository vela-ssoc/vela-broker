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
	"github.com/vela-ssoc/vela-broker/brkapi/internal/reqresp"
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
	rgb.Route("/mrr/:mid/*path").Any(sc.Mrr)
	rgb.Route("/mws/:mid/*path").GET(sc.Mws)
	rgb.Route("/brr/syscmd").GET(sc.Syscmd)
	rgb.Route("/brr/fm").GET(sc.FM)
	rgb.Route("/bws/echo").GET(sc.Echo)
}

func (sc *attachCtrl) Mrr(c *ship.Context) error {
	mid := c.Param("mid")
	path := c.Param("path")
	w, r := c.ResponseWriter(), c.Request()
	op := opurl.BMrr(mid, c.Method(), path, r.URL.RawQuery)
	sc.hub.Forward(op, w, r)

	return nil
}

func (sc *attachCtrl) Mws(c *ship.Context) error {
	mid := c.Param("mid")
	path := c.Param("path")
	w, r := c.ResponseWriter(), c.Request()
	query := r.URL.RawQuery
	op := opurl.BMws(mid, path, query)
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
	var req reqresp.Syscmd
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	timeout := req.Timeout
	if timeout < time.Second || timeout > time.Minute {
		timeout = 30 * time.Second
	}

	// 主要是增加用户的易用性
	cmd, args := req.Cmd, req.Args
	asz := len(args)
	if asz == 0 {
		split := strings.Split(cmd, " ")
		if len(split) > 1 {
			cmd, args = split[0], split[1:]
		}
	} else if asz == 1 {
		args = strings.Split(args[0], " ")
	}

	c.Warnf("命令执行: %s %s", cmd, strings.Join(args, " "))
	ret := syscmd.ExecTimeout(timeout, cmd, args...)

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
	for _, info := range infos {
		nm := info.Name()
		if match != "" {
			if matched, err := filepath.Match(match, nm); err == nil && !matched {
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
		}
	}

	return nil
}

type socketReply struct {
	Type   int       `json:"type"`
	Data   string    `json:"data"`
	TimeAt time.Time `json:"time_at"`
}
