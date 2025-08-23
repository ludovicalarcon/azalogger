package azalogger

import (
	"context"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger *zap.SugaredLogger
	level  *zap.AtomicLevel
}

func (l *zapLogger) Debug(msg string, kv ...any) { l.logger.Debugw(msg, kv...) }
func (l *zapLogger) Info(msg string, kv ...any)  { l.logger.Infow(msg, kv...) }
func (l *zapLogger) Warn(msg string, kv ...any)  { l.logger.Warnw(msg, kv...) }
func (l *zapLogger) Error(msg string, kv ...any) { l.logger.Errorw(msg, kv...) }
func (l *zapLogger) Fatal(msg string, kv ...any) { l.logger.Fatalw(msg, kv...) }

func (l *zapLogger) Sync() { _ = l.logger.Sync() }

func (l *zapLogger) With(kv ...any) Logger {
	return &zapLogger{logger: l.logger.With(kv...)}
}

func (l *zapLogger) WithContext(ctx context.Context) Logger {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	if !spanCtx.IsValid() {
		return l
	}

	return l.With("trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String())
}

func (l *zapLogger) HTTPLevelHandler(authHandler AuthorizationHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authHandler != nil && !authHandler(r) {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		l.level.ServeHTTP(w, r)
	})
}

func (l *zapLogger) LogLevel() string {
	return l.logger.Level().String()
}

func newZapLogger(cfg Config) (*zapLogger, error) {
	zapCfg := createZapConfig(cfg)
	logger, err := zapCfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &zapLogger{
		logger: logger.Sugar(),
		level:  &zapCfg.Level,
	}, nil
}

func createZapConfig(cfg Config) zap.Config {
	logLevel := getLogLevel(cfg)

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(logLevel)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var zapCfg zap.Config
	switch cfg.Env {
	case DevEnvironment:
		zapCfg = zap.NewDevelopmentConfig()
		zapCfg.Encoding = "console"
		zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		zapCfg = zap.NewProductionConfig()
		zapCfg.Encoding = "json"
	}

	zapCfg.Level = zap.NewAtomicLevelAt(zapLevel)
	zapCfg.EncoderConfig.TimeKey = "timestamp"
	zapCfg.EncoderConfig.EncodeTime = zapcore.TimeEncoder(func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.UTC().Format("2006-01-02T15:04:05Z0700"))
	})

	return zapCfg
}
