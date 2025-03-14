package agtapi

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/vela-ssoc/ssoc-broker/app/internal/param"
	"github.com/vela-ssoc/ssoc-broker/app/route"
	"github.com/vela-ssoc/ssoc-broker/bridge/mlink"
	"github.com/vela-ssoc/ssoc-common-mb/dal/gridfs"
	"github.com/vela-ssoc/ssoc-common-mb/dal/model"
	"github.com/vela-ssoc/ssoc-common-mb/dal/query"
	"github.com/vela-ssoc/vela-common-mba/ciphertext"
	"github.com/vela-ssoc/vela-common-mba/definition"
	"github.com/xgfone/ship/v5"
	"gorm.io/gen"
	"gorm.io/gorm"
)

func Upgrade(qry *query.Query, bid int64, gfs gridfs.FS) route.Router {
	return &upgradeREST{
		qry:     qry,
		bid:     bid,
		gfs:     gfs,
		maxsize: 200,
	}
}

type upgradeREST struct {
	qry     *query.Query
	bid     int64
	gfs     gridfs.FS
	mutex   sync.Mutex
	maxsize int
	count   int
}

func (rest *upgradeREST) Route(r *ship.RouteGroupBuilder) {
	r.Route("/broker/upgrade/download").Data("agent 下载二进制文件升级").GET(rest.Download)
}

func (rest *upgradeREST) Download(c *ship.Context) error {
	var req param.UpgradeDownload
	if err := c.BindQuery(&req); err != nil {
		return err
	}

	r := c.Request()
	ctx := r.Context()
	inf := mlink.Ctx(r.Context()) // 获取节点的信息
	ident := inf.Ident()
	if ident.Customized == req.Customized &&
		ident.Semver == req.Version {
		c.WriteHeader(http.StatusNotModified)
		return nil
	}

	if !rest.tryLock() {
		return c.NoContent(http.StatusTooManyRequests)
	}
	defer rest.unlock()

	bin, err := rest.matchBinary(ctx, inf, &req)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.WriteHeader(http.StatusNotModified)
			return nil
		}
		c.Warnf("查询更新版本出错：%s", err)
		return err
	}

	// 查询 broker 信息
	brkTbl := rest.qry.Broker
	brk, err := brkTbl.WithContext(ctx).Where(brkTbl.ID.Eq(rest.bid)).First()
	if err != nil {
		c.Warnf("更新版本查询 broker 信息错误：%s", err)
		return err
	}

	file, err := rest.gfs.OpenID(bin.FileID)
	if err != nil {
		c.Warnf("打开二进制文件错误：%s", err)
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer file.Close()

	addrs := make([]string, 0, 16)
	unique := make(map[string]struct{}, 16)
	for _, addr := range brk.LAN {
		if _, ok := unique[addr]; ok {
			continue
		}
		unique[addr] = struct{}{}
		addrs = append(addrs, addr)
	}
	for _, addr := range brk.VIP {
		if _, ok := unique[addr]; ok {
			continue
		}
		unique[addr] = struct{}{}
		addrs = append(addrs, addr)
	}

	semver := string(bin.Semver)
	hide := &definition.MHide{
		Servername: brk.Servername,
		Addrs:      addrs,
		Semver:     semver,
		Hash:       bin.Hash,
		Size:       bin.Size,
		Tags:       req.Tags,
		Goos:       bin.Goos,
		Arch:       bin.Arch,
		Unstable:   bin.Unstable,
		Customized: bin.Customized,
		DownloadAt: time.Now(),
		VIP:        brk.VIP,
		LAN:        brk.LAN,
		Edition:    semver,
	}

	enc, exx := ciphertext.EncryptPayload(hide)
	if exx != nil {
		return exx
	}
	stm := gridfs.Merge(file, enc)

	// 此时的 Content-Length = 原始文件 + 隐藏文件
	c.Header().Set(ship.HeaderContentLength, stm.ContentLength())
	c.Header().Set(ship.HeaderContentDisposition, stm.Disposition())

	return c.Stream(http.StatusOK, stm.ContentType(), stm)
}

func (rest *upgradeREST) tryLock() bool {
	rest.mutex.Lock()
	defer rest.mutex.Unlock()

	ok := rest.maxsize > rest.count
	if ok {
		rest.count++
	}

	return ok
}

func (rest *upgradeREST) unlock() {
	rest.mutex.Lock()
	defer rest.mutex.Unlock()
	rest.count--
	if rest.count < 0 {
		rest.count = 0
	}
}

func (rest *upgradeREST) matchBinary(ctx context.Context, inf mlink.Infer, req *param.UpgradeDownload) (*model.MinionBin, error) {
	tbl := rest.qry.MinionBin
	ident := inf.Ident()
	goos, arch := ident.Goos, ident.Arch
	conds := []gen.Condition{
		tbl.Deprecated.Is(false),
		tbl.Goos.Eq(goos),
		tbl.Arch.Eq(arch),
	}

	semver := req.Version
	if semver == "" { // 版本号为空说明是全局推送更新。
		if ident.Unstable { // 节点当前运行的是测试版，忽略更新指令。
			return nil, gorm.ErrRecordNotFound
		}
		// 按照节点当前的运行版本查找最新版本。
		weight := model.Semver(ident.Semver).Uint64() // 当前节点运行的版本。
		conds = append(conds, tbl.Weight.Gt(weight))
		conds = append(conds, tbl.Customized.Value(ident.Customized))
		conds = append(conds, tbl.Unstable.Is(false))
		return tbl.WithContext(ctx).
			Where(conds...).
			Order(tbl.Weight.Desc(), tbl.Semver.Desc()).
			First()
	}

	conds = append(conds, tbl.Semver.Eq(semver))
	conds = append(conds, tbl.Customized.Eq(req.Customized))

	return tbl.WithContext(ctx).Where(conds...).First()
}
