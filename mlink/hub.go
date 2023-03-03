package mlink

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vela-ssoc/backend-common/logback"
	"github.com/vela-ssoc/backend-common/model"
	"github.com/vela-ssoc/backend-common/netutil"
	"github.com/vela-ssoc/backend-common/opurl"
	"github.com/vela-ssoc/backend-common/spdy"
	"github.com/vela-ssoc/vela-broker/dialmgt"
	"gorm.io/gorm"
)

type Huber interface {
	Joiner
	Hide() dialmgt.Hide
	Ident() dialmgt.Ident
	Issue() dialmgt.Issue
	NodeName() string
	ResetDB() error
	Forward(opurl.URLer, http.ResponseWriter, *http.Request)
	Stream(op opurl.URLer, header http.Header) (*websocket.Conn, error)
}

var (
	ErrMinionBadInet  = errors.New("minion IP 不合法")
	ErrMinionRegister = errors.New("节点正在注册")
	ErrMinionInactive = errors.New("节点未激活")
	ErrMinionRemove   = errors.New("节点已删除")
	ErrMinionOnline   = errors.New("节点已经在线")
	ErrMinionOffline  = errors.New("节点未在线")
)

func Hub(db *gorm.DB, link dialmgt.Linker, handle http.Handler, slog logback.Logger) Huber {
	if handle == nil {
		handle = http.DefaultServeMux
	}
	section := container()
	random := rand.New(rand.NewSource(time.Now().UnixNano()))

	hub := &minionHub{
		db:      db,
		random:  random,
		section: section,
		link:    link,
		slog:    slog,
		handle:  handle,
	}
	transport := &http.Transport{DialContext: hub.dialContext}
	hub.client = opurl.NewClient(transport, slog)
	node := link.NodeName()
	hub.proxy = netutil.Forward(transport, node)
	hub.stream = netutil.Stream(hub.dialContext)

	return hub
}

type minionHub struct {
	db      *gorm.DB
	random  *rand.Rand
	section subsection
	link    dialmgt.Linker
	handle  http.Handler
	slog    logback.Logger
	client  opurl.Client
	proxy   netutil.Forwarder
	stream  netutil.Streamer
}

func (hub *minionHub) Hide() dialmgt.Hide {
	return hub.link.Hide()
}

func (hub *minionHub) Ident() dialmgt.Ident {
	return hub.link.Ident()
}

func (hub *minionHub) Issue() dialmgt.Issue {
	return hub.link.Issue()
}

func (hub *minionHub) NodeName() string {
	return hub.link.NodeName()
}

func (hub *minionHub) Auth(ident Ident) (Issue, http.Header, bool, error) {
	var issue Issue
	ip := ident.Inet.To4()
	if ip == nil || ip.IsLoopback() || ip.IsUnspecified() {
		return issue, nil, false, ErrMinionBadInet
	}
	inet := ip.String()
	// 根据 inet 查询节点信息
	var mon model.Minion
	if err := hub.db.First(&mon, "inet = ?", inet).Error; err != nil {
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
			BrokerID:   hub.link.Ident().ID,
			BrokerName: hub.link.Issue().Name,
		}
		if err = hub.db.Create(join).Error; err != nil {
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
	psz := hub.random.Intn(33) + 32
	passwd := make([]byte, psz)
	hub.random.Read(passwd)

	issue.ID, issue.Passwd = mon.ID, passwd

	return issue, nil, false, nil
}

func (hub *minionHub) Join(tran net.Conn, ident Ident, issue Issue) error {
	mux := spdy.Server(tran, spdy.WithEncrypt(issue.Passwd))
	//goland:noinspection GoUnhandledErrorResult
	defer mux.Close()

	id := issue.ID
	sid := strconv.FormatInt(id, 10) // 方便 dialContext
	conn := &connect{
		id:    id,
		ident: ident,
		issue: issue,
		mux:   mux,
	}

	if !hub.section.put(sid, conn) {
		return ErrMinionOnline
	}
	defer hub.section.del(sid)

	inet := ident.Inet.String()
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
		BrokerID:   hub.link.Ident().ID,
		BrokerName: hub.link.Issue().Name,
	}
	if err := hub.db.UpdateColumns(mon).Error; err != nil {
		return err
	}
	defer func() {
		hub.db.Model(mon).
			Where("status = ?", model.MinionOnline).
			UpdateColumn("status", model.MinionOffline)
	}()

	hub.slog.Infof("minion 节点 %s 上线了", inet)
	srv := &http.Server{
		Handler: hub.handle,
		BaseContext: func(net.Listener) context.Context {
			return context.WithValue(context.Background(), minionCtxKey, conn)
		},
	}
	_ = srv.Serve(mux)
	hub.slog.Warnf("minion 节点 %s 下线了", inet)

	return nil
}

// ResetDB 将所有连接该 broker 的节点数据库状态改为离线
func (hub *minionHub) ResetDB() error {
	brk := hub.link.Ident()
	return hub.db.Model(&model.Minion{}).
		Where("broker_id = ? AND status = ?", brk.ID, model.MinionOnline).
		UpdateColumn("status", model.MinionOffline).
		Error
}

func (hub *minionHub) BrkHide() dialmgt.Hide {
	return hub.link.Hide()
}

func (hub *minionHub) BrkIdent() dialmgt.Ident {
	return hub.link.Ident()
}

func (hub *minionHub) BrkIssue() dialmgt.Issue {
	return hub.link.Issue()
}

func (hub *minionHub) BrkInet() net.IP {
	return hub.BrkIdent().Inet
}

func (hub *minionHub) Fetch(ctx context.Context, op opurl.URLer, body io.Reader) (*http.Response, error) {
	return hub.client.Fetch(ctx, op, nil, body)
}

func (hub *minionHub) JSON(ctx context.Context, op opurl.URLer, body, reply any) error {
	return hub.client.JSON(ctx, op, nil, body, reply)
}

func (hub *minionHub) Forward(op opurl.URLer, w http.ResponseWriter, r *http.Request) {
	hub.proxy.Forward(op, w, r)
}

func (hub *minionHub) Stream(op opurl.URLer, header http.Header) (*websocket.Conn, error) {
	addr := op.String()
	conn, _, err := hub.stream.Stream(addr, header)
	if err == nil {
		hub.slog.Infof("建立 stream (%s) 通道成功", addr)
	} else {
		hub.slog.Warnf("建立 stream (%s) 通道失败：%s", addr, err)
	}
	return conn, err
}

func (hub *minionHub) dialContext(_ context.Context, _, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if conn, _ := hub.section.get(host); conn != nil {
		return conn.mux.Dial()
	}

	return nil, ErrMinionOffline
}
