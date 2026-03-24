// Package logger provides logging implementations.
package logger

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/sereozha/finance-tracker/internal/ports/logger"
	"github.com/sereozha/finance-tracker/pkg/contextkey"
)

// ZapLogger implements logger.Logger using Uber's Zap.
type ZapLogger struct {
	zap   *zap.Logger
	sugar *zap.SugaredLogger
}

// NewZapLogger creates a new Zap-based logger.
func NewZapLogger(env string) (logger.Logger, error) {
	var config zap.Config

	if env == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.EncoderConfig.StacktraceKey = "stacktrace"
	config.EncoderConfig.CallerKey = "caller"
	config.DisableCaller = false

	zapLogger, err := config.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return &ZapLogger{
		zap:   zapLogger,
		sugar: zapLogger.Sugar(),
	}, nil
}

// NewNopLogger creates a no-op logger for testing.
func NewNopLogger() logger.Logger {
	return &ZapLogger{
		zap:   zap.NewNop(),
		sugar: zap.NewNop().Sugar(),
	}
}

func (l *ZapLogger) Debug(msg string, fields ...logger.Field) {
	l.zap.Debug(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Info(msg string, fields ...logger.Field) {
	l.zap.Info(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Warn(msg string, fields ...logger.Field) {
	l.zap.Warn(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Error(msg string, fields ...logger.Field) {
	l.zap.Error(msg, toZapFields(fields)...)
}

func (l *ZapLogger) Fatal(msg string, fields ...logger.Field) {
	l.zap.Fatal(msg, toZapFields(fields)...)
	os.Exit(1)
}

func (l *ZapLogger) With(fields ...logger.Field) logger.Logger {
	return &ZapLogger{
		zap:   l.zap.With(toZapFields(fields)...),
		sugar: l.sugar.With(toArgs(fields)...),
	}
}

func (l *ZapLogger) WithContext(ctx context.Context) logger.Logger {
	fields := []logger.Field{}

	if cid := contextkey.CorrelationID(ctx); cid != "" {
		fields = append(fields, logger.String(logger.FieldCorrelationID, cid))
	}
	if uid := contextkey.UserID(ctx); uid != "" {
		fields = append(fields, logger.String(logger.FieldUserID, uid))
	}

	return l.With(fields...)
}

func (l *ZapLogger) Sync() error {
	return l.zap.Sync()
}

func toZapFields(fields []logger.Field) []zap.Field {
	result := make([]zap.Field, len(fields))
	for i, f := range fields {
		switch v := f.Value.(type) {
		case string:
			result[i] = zap.String(f.Key, v)
		case int:
			result[i] = zap.Int(f.Key, v)
		case int64:
			result[i] = zap.Int64(f.Key, v)
		case float64:
			result[i] = zap.Float64(f.Key, v)
		case bool:
			result[i] = zap.Bool(f.Key, v)
		case error:
			result[i] = zap.Error(v)
		case time.Duration:
			result[i] = zap.Duration(f.Key, v)
		default:
			result[i] = zap.Any(f.Key, v)
		}
	}
	return result
}

func toArgs(fields []logger.Field) []interface{} {
	args := make([]interface{}, 0, len(fields)*2)
	for _, f := range fields {
		args = append(args, f.Key, f.Value)
	}
	return args
}
