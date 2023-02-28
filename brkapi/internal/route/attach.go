package route

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/vela-ssoc/backend-common/pubody"
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
	rgb.Route("/brr/fm").GET(sc.FM)
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
	// ret.Files.NameDesc()

	return c.JSON(http.StatusOK, ret)
}
