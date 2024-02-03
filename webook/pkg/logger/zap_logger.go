package logger

import "go.uber.org/zap"

type ZapLogger struct {
	l *zap.Logger
}

func (z *ZapLogger) Debug(msg string, args ...Field) {
	zap.L()
	z.l.Debug(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Info(msg string, args ...Field) {
	z.l.Info(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Warn(msg string, args ...Field) {
	z.l.Warn(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) Error(msg string, args ...Field) {
	z.l.Error(msg, z.toZapFields(args)...)
}

func (z *ZapLogger) toZapFields(args []Field) []zap.Field {
	ret := make([]zap.Field, 0, len(args))
	for _, arg := range args {
		ret = append(ret, zap.Any(arg.Key, arg.Value))
	}
	return ret
}

func NewZapLogger(l *zap.Logger) Logger {
	return &ZapLogger{
		l: l,
	}
}
