package logback

import (
	"fmt"
	"log"
	"os"
)

type Logger interface {
	Tracef(format string, args ...any)
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type Replacer interface {
	Logger
	Replace(Logger)
}

func New() Replacer {
	flag := log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	lg := log.New(os.Stdout, "", flag)
	std := &stdlog{out: lg}

	return &replaceable{logger: std}
}

type replaceable struct {
	logger Logger
}

func (r *replaceable) Replace(lg Logger) {
	r.logger = lg
}

func (r *replaceable) Tracef(format string, args ...any) {
	r.logger.Tracef(format, args...)
}

func (r *replaceable) Debugf(format string, args ...any) {
	r.logger.Tracef(format, args...)
}

func (r *replaceable) Infof(format string, args ...any) {
	r.logger.Infof(format, args...)
}

func (r *replaceable) Warnf(format string, args ...any) {
	r.logger.Warnf(format, args...)
}

func (r *replaceable) Errorf(format string, args ...any) {
	r.logger.Errorf(format, args...)
}

type stdlog struct {
	out *log.Logger
}

func (l stdlog) output(level, format string, args ...any) {
	if len(args) == 0 {
		_ = l.out.Output(4, level+format)
	} else {
		_ = l.out.Output(4, fmt.Sprintf(level+format, args...))
	}
}

func (l stdlog) Tracef(format string, args ...any) {
	l.output("[T] ", format, args...)
}

func (l stdlog) Debugf(format string, args ...any) {
	l.output("[D] ", format, args...)
}

func (l stdlog) Infof(format string, args ...any) {
	l.output("[I] ", format, args...)
}

func (l stdlog) Warnf(format string, args ...any) {
	l.output("[W] ", format, args...)
}

func (l stdlog) Errorf(format string, args ...any) {
	l.output("[E] ", format, args...)
}
