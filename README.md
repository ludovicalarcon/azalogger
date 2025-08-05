# 📦 aza-logger

**A generic logger interface for Go**, designed for structured logging, OpenTelemetry integration, and testability.

---

## ✨ Features

- ✅ Unified `Logger` interface
- ✅ Zap backend with structured, high-performance logs
- ✅ In-memory backend for test logging
- ✅ Context propagation with `WithContext(ctx)`
- ✅ Field injection with `With(...)`
- ✅ Optional HTTP log level control (via `/loglevel`)
- ✅ Designed for use with dependency injection or as a singleton

---

## 🚀 Getting Started

### 1. Initialization

```go
import (
 azalogger "gitlab.com/ludovic-alarcon/aza-logger"
)

func main() {
 log, err := azalogger.NewLogger(azalogger.Config{
  Backend:  azalogger.ZapBackend,
  Env:      azalogger.ProdEnvironment,
  LogLevel: azalogger.InfoLevel,
 })
 if err != nil {
  panic(err)
 }

// Field injection on all logs
 log = log.With("app", "my-service")
 log.Info("service started")
}
```

### 2. 🔄 Runtime Log Level Control

If using a backend that support dynamic log level, you can expose a log level handler:

```go
http.Handle("/loglevel", log.HTTPLevelHandler())
go http.ListenAndServe(":8080", nil)

```

Then use:

```go
curl localhost:8080/loglevel
curl -X PUT localhost:8080/loglevel -d '{"level":"debug"}'
```

Backends that don’t support dynamic log level return 501.

Backends with support

- ✅ Zap

### 3. In-memory logger

The in-memory logger implementation is perfect to be used in unit test.
Just need to call the Entries method to get a slice of log

```go
log, err := azalogger.NewLogger(azalogger.Config{
 Backend:  azalogger.ZapBackend,
 Env:      azalogger.ProdEnvironment,
 LogLevel: azalogger.InfoLevel,
})
 
if err != nil {
  panic(err)
}

log.Info("service started")
logs := log.(*azalogger.InMemoryLogger).Entries()
```
