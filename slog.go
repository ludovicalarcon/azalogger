package azalogger

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"

	"go.opentelemetry.io/otel/trace"
)

type slogLogger struct {
	logger *slog.Logger
	level  *slog.LevelVar
}

func (l *slogLogger) Debug(msg string, kv ...any) { l.logger.Debug(msg, kv...) }
func (l *slogLogger) Info(msg string, kv ...any)  { l.logger.Info(msg, kv...) }
func (l *slogLogger) Warn(msg string, kv ...any)  { l.logger.Warn(msg, kv...) }
func (l *slogLogger) Error(msg string, kv ...any) { l.logger.Error(msg, kv...) }
func (l *slogLogger) Fatal(msg string, kv ...any) {
	l.logger.Error(msg, kv...)
	os.Exit(1)
}

// Sync -> NOOP
func (l *slogLogger) Sync() {}

func (l *slogLogger) With(kv ...any) Logger {
	return &slogLogger{logger: l.logger.With(kv...)}
}

func (l *slogLogger) WithContext(ctx context.Context) Logger {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	if !spanCtx.IsValid() {
		return l
	}

	return l.With("trace_id", spanCtx.TraceID().String(),
		"span_id", spanCtx.SpanID().String())
}

func (l *slogLogger) HTTPLevelHandler(authHandler AuthorizationHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authHandler != nil && !authHandler(r) {
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		switch r.Method {
		case http.MethodGet:
			level := l.level.Level().String()
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"level": level})
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			var payload struct {
				Level string `json:"level"`
			}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, "invalid payload", http.StatusBadRequest)
				return
			}

			newLevel, err := parseSlogLevel(payload.Level)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			l.level.Set(newLevel)
			w.WriteHeader(http.StatusOK)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})
}

func parseSlogLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error", "fatal":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, errors.New("invalid log level")
	}
}

func (l *slogLogger) LogLevel() string {
	return l.level.Level().String()
}

func newSlogLogger(cfg Config) *slogLogger {
	logLevel, err := parseSlogLevel(getLogLevel(cfg).String())
	if err != nil {
		logLevel = slog.LevelInfo
	}

	level := &slog.LevelVar{}
	level.Set(logLevel)
	switch cfg.Env {
	case DevEnvironment:
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		logger := slog.New(handler)

		return &slogLogger{
			logger: logger,
			level:  level,
		}
	default:
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		logger := slog.New(handler)

		return &slogLogger{
			logger: logger,
			level:  level,
		}
	}
}
