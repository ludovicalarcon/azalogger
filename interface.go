// Package azalogger defines a generic logging interface.
//
// It supports structured logging with key/value pairs, log levels, and context propagation.
// The interface is designed to be backend-agnostic and works with defined implementations (Backend type below)
// Use NewLogger(cfg) to instantiate a logger. Then inject it into your application components.
//
// Example:
//
//	log, err := logger.NewLogger(logger.Config{
//	  Backend: logger.ZapBackend,
//	  Env:     logger.ProdEnvironment,
//	  Level:   logger.InfoLevel,
//	})
//
//	if err != nil {
//	  panic(err)
//	}
//
//	log = log.With("app", "my-service")
//	log.Info("startup complete")
//
// Backends may optionally support dynamic log level changes via HTTP with HTTPLevelHandler().
package azalogger

import (
	"context"
	"net/http"
	"os"
)

type (
	LogLevel    string
	Environment string
	Backend     int
)

const (
	ZapBackend Backend = iota
	SlogBackend
	InMemoryBackend
)

const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"

	DevEnvironment  Environment = "dev"
	ProdEnvironment Environment = "prod"

	LogLevelEnvVar = "AZA_LOG_LEVEL"
)

type Config struct {
	LogLevel LogLevel
	Env      Environment
	Backend  Backend
}

// Handler to check if the request is allowed to modify log level
type AuthorizationHandler func(r *http.Request) bool

func (l LogLevel) String() string {
	return string(l)
}

type Logger interface {
	Debug(msg string, keysAndValues ...any)
	Info(msg string, keysAndValues ...any)
	Warn(msg string, keysAndValues ...any)
	Error(msg string, keysAndValues ...any)
	Fatal(msg string, keysAndValues ...any)
	Sync()

	With(keysAndValues ...any) Logger
	WithContext(ctx context.Context) Logger

	// HTTP handler to change loglevel at runtime
	HTTPLevelHandler(authHandler AuthorizationHandler) http.Handler

	LogLevel() string
}

func getLogLevel(cfg Config) LogLevel {
	level := LogLevel(os.Getenv(LogLevelEnvVar))
	if level == "" {
		level = cfg.LogLevel
		if level == "" {
			level = InfoLevel
		}
	}
	return level
}
