package mlink

import (
	"context"
	"database/sql"
	"errors"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/model"
	"github.com/vela-ssoc/backend-common/spdy"
	"github.com/vela-ssoc/vela-broker/dialmgt"
	"gorm.io/gorm"
)

type Huber interface {
	Joiner
	ResetDB() error
}

var (
	ErrMinionBadInet  = errors.New("minion IP 不合法")
	ErrMinionRegister = errors.New("节点正在注册")
	ErrMinionInactive = errors.New("节点未激活")
	ErrMinionRemove   = errors.New("节点已删除")
	ErrMinionOnline   = errors.New("节点已经在线")
)

func Hub(db *gorm.DB, link dialmgt.Linker, handle http.Handler, slog logback.Logger) Huber {
	if handle == nil {
		handle = http.DefaultServeMux
	}
	section := container()
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &hub{
		db:      db,
		random:  random,
		section: section,
		link:    link,
		slog:    slog,
		handle:  handle,
	}
}

type hub struct {
	db      *gorm.DB
	random  *rand.Rand
	section subsection
	link    dialmgt.Linker
	handle  http.Handler
	slog    logback.Logger
}

func (hb *hub) Auth(ident Ident) (Issue, http.Header, bool, error) {
	var issue Issue
	ip := ident.IP.To4()
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
		return issue, nil, false, ErrMinionBadInet
	}
	inet := ip.String()
	// 根据 inet 查询节点信息
	var mon model.Minion
	if err := hb.db.First(&mon, "inet = ?", inet).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return issue, nil, false, err
		}

		now := sql.NullTime{Valid: true, Time: time.Now()}
		join := &model.Minion{
			Inet:       inet,
			Name:       inet,
			MAC:        ident.MAC,
			Goos:       ident.Goos,
			Arch:       ident.Arch,
			Semver:     ident.Semver,
			CPU:        ident.CPU,
			PID:        ident.PID,
			Username:   ident.Username,
			Hostname:   ident.Hostname,
			Workdir:    ident.Workdir,
			Executable: ident.Executable,
			JoinedAt:   now,
			BrokerID:   hb.link.Ident().ID,
			BrokerName: hb.link.Issue().Name,
		}
		if err = hb.db.Create(join).Error; err != nil {
			return issue, nil, false, err
		}
		return issue, nil, false, ErrMinionRegister
	}

	status := mon.Status
	if status == model.MinionInactive {
		return issue, nil, false, ErrMinionInactive
	}
	if status == model.MinionRemove {
		return issue, nil, true, ErrMinionRemove
	}
	if status == model.MinionOnline {
		return issue, nil, true, ErrMinionOnline
	}

	// 随机生成一个 32-64 位长度的加密密钥
	psz := hb.random.Intn(33) + 32
	passwd := make([]byte, psz)
	hb.random.Read(passwd)

	issue.ID, issue.Passwd = mon.ID, passwd

	return issue, nil, false, nil
}

func (hb *hub) Join(tran net.Conn, ident Ident, issue Issue) error {
	mux := spdy.Server(tran, spdy.WithEncrypt(issue.Passwd))
	//goland:noinspection GoUnhandledErrorResult
	defer mux.Close()

	conn := &connect{
		ident: ident,
		issue: issue,
		mux:   mux,
	}

	id := issue.ID
	if !hb.section.put(id, conn) {
		return ErrMinionOnline
	}
	defer hb.section.del(id)

	inet := ident.IP.String()
	now := sql.NullTime{Valid: true, Time: time.Now()}
	mon := &model.Minion{
		ID:         id,
		Inet:       inet,
		Status:     model.MinionOnline,
		MAC:        ident.MAC,
		Goos:       ident.Goos,
		Arch:       ident.Arch,
		Semver:     ident.Semver,
		CPU:        ident.CPU,
		PID:        ident.PID,
		Username:   ident.Username,
		Hostname:   ident.Hostname,
		Workdir:    ident.Workdir,
		Executable: ident.Executable,
		PingedAt:   now,
		JoinedAt:   now,
		BrokerID:   hb.link.Ident().ID,
		BrokerName: hb.link.Issue().Name,
	}
	if err := hb.db.UpdateColumns(mon).Error; err != nil {
		return err
	}
	defer func() {
		hb.db.Model(mon).
			Where("status = ?", model.MinionOnline).
			UpdateColumn("status", model.MinionOffline)
	}()

	hb.slog.Infof("minion 节点 %s 上线了", inet)
	srv := &http.Server{
		Handler: hb.handle,
		BaseContext: func(net.Listener) context.Context {
			return context.WithValue(context.Background(), minionCtxKey, conn)
		},
	}
	_ = srv.Serve(mux)
	hb.slog.Warnf("minion 节点 %s 下线了", inet)

	return nil
}

// ResetDB 将所有连接该 broker 的节点数据库状态改为离线
func (hb *hub) ResetDB() error {
	brk := hb.link.Ident()
	return hb.db.Model(&model.Minion{}).
		Where("broker_id = ? AND status = ?", brk.ID, model.MinionOnline).
		UpdateColumn("status", model.MinionOffline).
		Error
}
