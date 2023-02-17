package logback

import (
	"context"
	"errors"
	"runtime"
	"strings"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func GORM(l *zap.Logger, level string) logger.Interface {
	lvl := logger.Error
	switch strings.ToLower(level) {
	case "info":
		lvl = logger.Info
	case "warn":
		lvl = logger.Warn
	case "error":
		lvl = logger.Error
	}

	return &gormLog{
		zlog:                      l,
		level:                     lvl,
		slowThreshold:             200 * time.Millisecond,
		skipCallerLookup:          false,
		ignoreRecordNotFoundError: true,
	}
}

type gormLog struct {
	zlog                      *zap.Logger
	level                     logger.LogLevel
	slowThreshold             time.Duration
	skipCallerLookup          bool
	ignoreRecordNotFoundError bool
}

func (gl *gormLog) Tracef(tmpl string, args ...any) {
	if gl.level >= logger.Info {
		gl.zlog.Sugar().Infof(tmpl, args...)
	}
}

func (gl *gormLog) Debugf(tmpl string, args ...any) {
	if gl.level >= logger.Info {
		gl.zlog.Sugar().Infof(tmpl, args...)
	}
}

func (gl *gormLog) Infof(tmpl string, args ...any) {
	if gl.level >= logger.Info {
		gl.zlog.Sugar().Infof(tmpl, args...)
	}
}

func (gl *gormLog) Warnf(tmpl string, args ...any) {
	if gl.level >= logger.Warn {
		gl.zlog.Sugar().Warnf(tmpl, args...)
	}
}

func (gl *gormLog) Errorf(tmpl string, args ...any) {
	if gl.level >= logger.Error {
		gl.zlog.Sugar().Errorf(tmpl, args...)
	}
}

func (gl *gormLog) LogMode(level logger.LogLevel) logger.Interface {
	return &gormLog{
		zlog:                      gl.zlog,
		level:                     level,
		slowThreshold:             gl.slowThreshold,
		skipCallerLookup:          gl.skipCallerLookup,
		ignoreRecordNotFoundError: gl.ignoreRecordNotFoundError,
	}
}

func (gl *gormLog) Info(_ context.Context, str string, args ...interface{}) {
	if gl.level < logger.Info {
		return
	}
	gl.logger().Sugar().Infof(str, args...)
}

func (gl *gormLog) Warn(_ context.Context, str string, args ...interface{}) {
	if gl.level < logger.Warn {
		return
	}
	gl.logger().Sugar().Warnf(str, args...)
}

func (gl *gormLog) Error(_ context.Context, str string, args ...interface{}) {
	if gl.level < logger.Error {
		return
	}
	gl.logger().Sugar().Errorf(str, args...)
}

func (gl *gormLog) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if gl.level < logger.Silent {
		return
	}
	elapsed := time.Since(begin)
	switch {
	case err != nil && gl.level >= logger.Error && (!gl.ignoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		gl.logger().Sugar().Errorf("[elapsed %s, rows %d] %s", elapsed, rows, sql)
	case gl.slowThreshold != 0 && elapsed > gl.slowThreshold && gl.level >= logger.Warn:
		sql, rows := fc()
		if err != nil {
			gl.logger().Sugar().Warnf("[elapsed %s, rows %d] %s, error: %s", elapsed, rows, sql, err.Error())
		} else {
			gl.logger().Sugar().Warnf("[elapsed %s, rows %d] %s", elapsed, rows, sql)
		}
	case gl.level >= logger.Info:
		sql, rows := fc()
		if err != nil {
			gl.logger().Sugar().Infof("[elapsed %s, rows %d] %s, error: %s", elapsed, rows, sql, err.Error())
		} else {
			gl.logger().Sugar().Infof("[elapsed %s, rows %d] %s", elapsed, rows, sql)
		}
	}
}

func (gl *gormLog) logger() *zap.Logger {
	for i := 2; i < 8; i++ {
		_, file, _, ok := runtime.Caller(i)
		switch {
		case !ok:
		case strings.HasSuffix(file, "_test.go"):
		case strings.Contains(file, "gorm.io/gorm"):
		default:
			return gl.zlog.WithOptions(zap.AddCallerSkip(i - 1))
		}
	}
	return gl.zlog
}
