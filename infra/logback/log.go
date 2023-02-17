package logback

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Trace(...any)
	Debug(...any)
	Info(...any)
	Warn(...any)
	Error(...any)
	Tracef(string, ...any)
	Debugf(string, ...any)
	Infof(string, ...any)
	Warnf(string, ...any)
	Errorf(string, ...any)
	Level() zapcore.Level
	Replace(*zap.Logger)
}

func Stdout() Logger {
	prod := zap.NewProductionEncoderConfig()
	prod.EncodeTime = zapcore.ISO8601TimeEncoder
	prod.EncodeLevel = zapcore.CapitalColorLevelEncoder

	encoder := zapcore.NewConsoleEncoder(prod)
	syncer := zapcore.AddSync(os.Stdout)
	level := zapcore.DebugLevel

	core := zapcore.NewCore(encoder, syncer, level)

	opts := []zap.Option{
		zap.WithCaller(true),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.AddCallerSkip(1),
	}
	lg := zap.New(core, opts...)

	return &sugarLog{sugar: lg.Sugar()}
}

type sugarLog struct {
	sugar *zap.SugaredLogger
}

func (sg *sugarLog) Trace(v ...any)            { sg.sugar.Debug(v...) }
func (sg *sugarLog) Debug(v ...any)            { sg.sugar.Debug(v...) }
func (sg *sugarLog) Info(v ...any)             { sg.sugar.Info(v...) }
func (sg *sugarLog) Warn(v ...any)             { sg.sugar.Warn(v...) }
func (sg *sugarLog) Error(v ...any)            { sg.sugar.Error(v...) }
func (sg *sugarLog) Tracef(s string, v ...any) { sg.sugar.Debugf(s, v...) }
func (sg *sugarLog) Debugf(s string, v ...any) { sg.sugar.Debugf(s, v...) }
func (sg *sugarLog) Infof(s string, v ...any)  { sg.sugar.Infof(s, v...) }
func (sg *sugarLog) Warnf(s string, v ...any)  { sg.sugar.Warnf(s, v...) }
func (sg *sugarLog) Errorf(s string, v ...any) { sg.sugar.Errorf(s, v...) }
func (sg *sugarLog) Level() zapcore.Level      { return sg.sugar.Level() }
func (sg *sugarLog) Replace(l *zap.Logger)     { sg.sugar = l.Sugar() }
