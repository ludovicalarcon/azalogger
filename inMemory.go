package azalogger

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// InMemoryLogger to be used for unit test
// Call Buffer() of concret type to get the content of in-memory logs
type InMemoryLogger struct {
	buffer         *bytes.Buffer
	mu             sync.Mutex
	logLevel       LogLevel
	injectedFields []string
}

func newInMemoryLogger(cfg Config) *InMemoryLogger {
	buffer := new(bytes.Buffer)
	buffer.Grow(1024)

	if !isValidLogLevel(cfg.LogLevel.String()) {
		cfg.LogLevel = InfoLevel
	}

	return &InMemoryLogger{
		buffer:         buffer,
		logLevel:       cfg.LogLevel,
		injectedFields: make([]string, 0, 2),
	}
}

func isValidLogLevel(logLevel string) bool {
	switch logLevel {
	case DebugLevel.String(), InfoLevel.String(), WarnLevel.String(),
		ErrorLevel.String(), FatalLevel.String():
		return true
	default:
		return false
	}
}

func (l *InMemoryLogger) Debug(msg string, kv ...any) {
	if l.logLevel == DebugLevel {
		l.log("DEBUG", msg, kv...)
	}
}

func (l *InMemoryLogger) Info(msg string, kv ...any) {
	if l.logLevel == DebugLevel || l.logLevel == InfoLevel {
		l.log("INFO", msg, kv...)
	}
}

func (l *InMemoryLogger) Warn(msg string, kv ...any) {
	if l.logLevel == DebugLevel ||
		l.logLevel == InfoLevel ||
		l.logLevel == WarnLevel {
		l.log("WARN", msg, kv...)
	}
}

func (l *InMemoryLogger) Error(msg string, kv ...any) {
	if l.logLevel == DebugLevel ||
		l.logLevel == InfoLevel ||
		l.logLevel == WarnLevel ||
		l.logLevel == ErrorLevel {
		l.log("ERROR", msg, kv...)
	}
}

func (l *InMemoryLogger) Fatal(msg string, kv ...any) { l.log("FATAL", msg, kv...) }

// Sync -> NOOP
func (l *InMemoryLogger) Sync() {}

func (l *InMemoryLogger) log(level, msg string, kv ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	fmt.Fprintf(l.buffer, "[%s] %s", level, msg)
	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			fmt.Fprintf(l.buffer, " %v=%v", kv[i], kv[i+1])
		}
	}
	for _, field := range l.injectedFields {
		fmt.Fprintf(l.buffer, " %s", field)
	}
	l.buffer.WriteByte('\n')
}

func (l *InMemoryLogger) With(kv ...any) Logger {
	l.mu.Lock()
	defer l.mu.Unlock()

	for i := 0; i < len(kv); i += 2 {
		if i+1 < len(kv) {
			l.injectedFields = append(l.injectedFields, fmt.Sprintf("%v=%v", kv[i], kv[i+1]))
		}
	}
	return l
}

func (l *InMemoryLogger) WithContext(ctx context.Context) Logger {
	return l
}

func (l *InMemoryLogger) HTTPLevelHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "log level control not supported for in-memory logger", http.StatusNotImplemented)
	})
}

func (l *InMemoryLogger) LogLevel() string {
	return l.logLevel.String()
}

// Entries is not part of interface
// only on concret type in-memory logger
// allows to get all in-memory logs
func (l *InMemoryLogger) Entries() []string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Split(l.buffer.String(), "\n")
}
