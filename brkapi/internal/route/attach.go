package route

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vela-ssoc/backend-common/pubrr"
	"github.com/vela-ssoc/backend-common/syscmd"
	"github.com/vela-ssoc/vela-broker/brkapi/internal/reqresp"
	"github.com/xgfone/ship/v5"
)

func Attach() RegRouter {
	return new(attachCtrl)
}

type attachCtrl struct{}

func (sc *attachCtrl) RegRoute(rgb *ship.RouteGroupBuilder) {
	rgb.Route("/brr/syscmd").GET(sc.Syscmd)
	rgb.Route("/brr/fs").GET(sc.FS)
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

func (sc *attachCtrl) FS(c *ship.Context) error {
	str := c.Query("path", "/")
	name := filepath.Clean(str)
	open, err := os.Open(name)
	if err != nil {
		return err
	}
	if stat, err := open.Stat(); err != nil {
		return err
	} else if !stat.IsDir() {
		return c.Attachment(name, "")
	}

	infos, err := open.Readdir(-1)
	_ = open.Close()
	if err != nil {
		return err
	}

	ret := &pubrr.FS{Abs: name}
	for _, info := range infos {
		nm := info.Name()
		fl := &pubrr.File{
			Path:  filepath.Join(name, nm),
			Name:  nm,
			Size:  info.Size(),
			Mtime: info.ModTime(),
			Dir:   info.IsDir(),
			Mode:  info.Mode().String(),
		}
		ret.Files = append(ret.Files, fl)
	}
	// ret.Files.NameDesc()

	return c.JSON(http.StatusOK, ret)
}
