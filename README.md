# 📦 aza-logger

**A generic logger interface for Go**, designed for structured logging, OpenTelemetry integration, and testability.

---

## ✨ Features

- ✅ Unified `Logger` interface
- ✅ Zap backend with structured, high-performance logs
- ✅ Slog backend with structured logs (stdlib)
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
  // Use azalogger.SlogBackend for slog logger
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

If using a backend that support dynamic log level, you can expose a log level HTTP handler with optional authorization:

```go
// Example: protect endpoint with simple API key
apiKey := "supersecretkey"
auth := func(r *http.Request) bool {
    return r.Header.Get("X-API-Key") == apiKey
}

// Expose log level endpoint
http.Handle("/loglevel", log.HTTPLevelHandler(auth))
go http.ListenAndServe(":8080", nil)

```

Then use:

```bash
# Get current log level (GET)
curl -H "X-API-Key: supersecretkey" localhost:8080/loglevel

# Change log level to debug (PUT)
curl -X PUT -H "X-API-Key: supersecretkey" \
     -H "Content-Type: application/json" \
     -d '{"level":"debug"}' \
     localhost:8080/loglevel
```

Requests without valid authorization will return `403 Forbidden`.

Backends that don’t support dynamic log level return `501 Not Implemented`.

##### Security recommendation

Always protect the log level endpoint in production  
attackers can silence logging or elevate debug verbosity

##### Backends with support

- ✅ Zap
- ✅ Slog

### 3. In-memory logger

The in-memory logger implementation is perfect to be used in unit test.
Just need to call the Entries method to get a slice of logs.
It's preferred to not use factory and call directly NewInMemoryLogger to leverage full features.
Some helpers are not part of the interface but useful for unit test, so better to instantiate
concrete type and inject it as interface type

```go
cfg := azalogger.Config{
 Backend:  azalogger.InMemoryBackend,
 Env:      azalogger.ProdEnvironment,
 LogLevel: azalogger.InfoLevel,
}
log := NewInMemoryLogger(cfg)
 
log.Info("service started")
logs := log.Entries()
```
