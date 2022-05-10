package logger

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debug(msg string, keyValueContext ...interface{})
	Info(msg string, keyValueContext ...interface{})
	Warn(msg string, keyValueContext ...interface{})
	Error(msg string, keyValueContext ...interface{})
}

var _ Logger = (*ZapLogger)(nil)

type ZapLogger struct {
	zl *zap.SugaredLogger
}

func New(level string, target string, encoding string) (*ZapLogger, error) {
	config := zap.NewProductionConfig()

	zlevel, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, fmt.Errorf("unknown level %s: %w", level, err)
	}

	config.Encoding = encoding
	config.EncoderConfig.EncodeTime = func(t time.Time, encoder zapcore.PrimitiveArrayEncoder) {
		formatted := t.Format(time.RFC3339)
		encoder.AppendString(formatted)
	}
	config.Level = zlevel
	config.OutputPaths = []string{target}
	config.ErrorOutputPaths = []string{target}
	config.DisableCaller = true
	config.DisableStacktrace = false

	zl, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("cannot build logger: %w", err)
	}

	return &ZapLogger{zl.Sugar()}, nil
}

func (l *ZapLogger) Debug(msg string, keyValueContext ...interface{}) {
	l.zl.Debugw(msg, keyValueContext...)
}

func (l *ZapLogger) Info(msg string, keyValueContext ...interface{}) {
	l.zl.Infow(msg, keyValueContext...)
}

func (l *ZapLogger) Warn(msg string, keyValueContext ...interface{}) {
	l.zl.Warnw(msg, keyValueContext...)
}

func (l *ZapLogger) Error(msg string, keyValueContext ...interface{}) {
	l.zl.Errorw(msg, keyValueContext...)
}

func (l *ZapLogger) Close() error {
	return l.zl.Sync()
}
