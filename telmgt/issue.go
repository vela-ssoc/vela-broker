package telmgt

import (
	"crypto/tls"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/vela-ssoc/backend-common/encipher"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Issue 认证成功后返回的必要信息
type Issue struct {
	Name     string   `json:"name"`     // 节点名字
	Passwd   []byte   `json:"passwd"`   // 通信加密密钥
	Listen   Listen   `json:"listen"`   // 服务监听配置
	Logger   Logger   `json:"logger"`   // 日志配置
	Database Database `json:"database"` // 数据库配置
}

// decrypt 解密消息
func (issue *Issue) decrypt(data []byte) error {
	return encipher.DecryptJSON(data, issue)
}

func (issue Issue) String() string {
	dat, _ := json.MarshalIndent(issue, "", "    ")
	return string(dat)
}

// Listen 本地服务监听配置
type Listen struct {
	Addr string `json:"addr"` // 监听地址 :8080 192.168.1.2:8080
	Cert []byte `json:"cert"` // 证书
	Pkey []byte `json:"pkey"` // 私钥
}

// Certifier 初始化证书
func (ln Listen) Certifier() ([]tls.Certificate, error) {
	if len(ln.Cert) == 0 || len(ln.Pkey) == 0 {
		return nil, nil
	}

	cert, err := tls.X509KeyPair(ln.Cert, ln.Pkey)
	if err != nil {
		return nil, err
	}

	return []tls.Certificate{cert}, nil
}

type Logger struct {
	Level     string `json:"level"     yaml:"level"`
	Console   bool   `json:"console"   yaml:"console"`
	Colorful  bool   `json:"colorful"  yaml:"colorful"`
	Directory string `json:"directory" yaml:"directory"`
	Maxsize   int    `json:"maxsize"   yaml:"maxsize"`
	MaxAge    int    `json:"maxage"    yaml:"maxage"`
	Backup    int    `json:"backup"    yaml:"backup"`
	Localtime bool   `json:"localtime" yaml:"localtime"`
	Compress  bool   `json:"compress"  yaml:"compress"`
}

func (l Logger) Zap() *zap.Logger {
	console := l.Console
	var filename string
	if dir := l.Directory; dir != "" {
		filename = filepath.Join(dir, "broker.log")
	}
	// 既不输出到控制台又不输出到日志文件
	if !console && filename == "" {
		return zap.NewNop()
	}

	prod := zap.NewProductionEncoderConfig()
	prod.EncodeTime = zapcore.ISO8601TimeEncoder
	if l.Colorful {
		prod.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		prod.EncodeLevel = zapcore.CapitalLevelEncoder
	}

	var syncer zapcore.WriteSyncer
	if console {
		syncer = zapcore.AddSync(os.Stdout)
	}
	if filename != "" {
		lumber := &lumberjack.Logger{
			Filename:   filename,
			MaxSize:    l.Maxsize,
			MaxAge:     l.MaxAge,
			MaxBackups: l.Backup,
			LocalTime:  l.Localtime,
			Compress:   l.Compress,
		}
		if syncer == nil {
			syncer = zapcore.AddSync(lumber)
		} else {
			syncer = zapcore.NewMultiWriteSyncer(syncer, zapcore.AddSync(lumber))
		}
	}

	encoder := zapcore.NewConsoleEncoder(prod)
	level := zapcore.WarnLevel
	_ = level.Set(l.Level) // 就算设置失败还是默认值 WarnLevel
	core := zapcore.NewCore(encoder, syncer, level)

	return zap.New(core, zap.WithCaller(true), zap.AddStacktrace(zapcore.ErrorLevel))
}

type Database struct {
	Level       string            `json:"level"         yaml:"level"`                                  // SQL 日志打印级别
	CDN         string            `json:"cdn"           yaml:"cdn"`                                    // 文件缓存位置
	MaxOpenConn int               `json:"max_open_conn" yaml:"max_open_conn"`                          // 最大连接数
	MaxIdleConn int               `json:"max_idle_conn" yaml:"max_idle_conn"`                          // 最大空闲连接数
	MaxLifeTime time.Duration     `json:"max_life_time" yaml:"max_life_time"`                          // 连接最大存活时长
	MaxIdleTime time.Duration     `json:"max_idle_time" yaml:"max_idle_time"`                          // 空闲连接最大时长
	DSN         string            `json:"dsn"           yaml:"dsn"`                                    // 数据源
	User        string            `json:"user"          yaml:"user"   validate:"required_without=DSN"` // 数据库用户名
	Passwd      string            `json:"passwd"        yaml:"passwd" validate:"required_without=DSN"` // 密码
	Net         string            `json:"net"           yaml:"net"`                                    // 连接协议
	Addr        string            `json:"addr"          yaml:"addr"   validate:"required_without=DSN"` // 连接地址
	DBName      string            `json:"dbname"        yaml:"dbname" validate:"required_without=DSN"` // 库名
	Params      map[string]string `json:"params"        yaml:"params"`                                 // 参数
}

// FormatDSN formats the given Config into a DSN string
// which can be passed to the driver.
func (db Database) FormatDSN() string {
	if dsn := db.DSN; dsn != "" {
		return dsn
	}

	protocol := db.Net
	if protocol == "" {
		protocol = "tcp" // 不填写则默认 tcp 协议
	}
	cfg := &mysql.Config{
		User:   db.User,
		Passwd: db.Passwd,
		Net:    protocol,
		Addr:   db.Addr,
		DBName: db.DBName,
		Params: db.Params,
	}

	return cfg.FormatDSN()
}
