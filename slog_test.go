package azalogger

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSlogLogger(t *testing.T) {
	t.Run("should create slog logger based on config (dev)", func(t *testing.T) {
		level := &slog.LevelVar{}
		level.Set(slog.LevelWarn)
		expected := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		cfg := Config{
			LogLevel: WarnLevel,
			Env:      DevEnvironment,
		}

		got := newSlogLogger(cfg)
		assert.Equal(t, expected, got.logger.Handler())
	})

	t.Run("should create slog logger based on config (prod)", func(t *testing.T) {
		level := &slog.LevelVar{}
		level.Set(slog.LevelWarn)
		expected := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		cfg := Config{
			LogLevel: WarnLevel,
			Env:      ProdEnvironment,
		}

		got := newSlogLogger(cfg)
		assert.Equal(t, expected, got.logger.Handler())
	})

	t.Run("should default to prod config and info loglevel when nothing is set", func(t *testing.T) {
		level := &slog.LevelVar{}
		level.Set(slog.LevelInfo)
		expected := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		cfg := Config{}

		got := newSlogLogger(cfg)
		assert.Equal(t, expected, got.logger.Handler())
	})

	t.Run("should default to prod config and info loglevel when unknown is set", func(t *testing.T) {
		t.Setenv(LogLevelEnvVar, "unknown")

		level := &slog.LevelVar{}
		level.Set(slog.LevelInfo)
		expected := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
		cfg := Config{}

		got := newSlogLogger(cfg)
		assert.Equal(t, expected, got.logger.Handler())
	})
}

func TestParseLogLevel(t *testing.T) {
	testcases := []struct {
		name     string
		level    string
		expected slog.Level
		onError  bool
	}{
		{
			name:     "should return debug",
			level:    DebugLevel.String(),
			expected: slog.LevelDebug,
			onError:  false,
		},
		{
			name:     "should return info",
			level:    InfoLevel.String(),
			expected: slog.LevelInfo,
			onError:  false,
		},
		{
			name:     "should return warn",
			level:    WarnLevel.String(),
			expected: slog.LevelWarn,
			onError:  false,
		},
		{
			name:     "should return error",
			level:    ErrorLevel.String(),
			expected: slog.LevelError,
			onError:  false,
		},
		{
			name:     "should return error for fatal",
			level:    FatalLevel.String(),
			expected: slog.LevelError,
			onError:  false,
		},
		{
			name:     "should return error on unknown",
			level:    "foo",
			expected: slog.LevelInfo,
			onError:  true,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseSlogLevel(tc.level)

			if tc.onError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestSlogLogs(t *testing.T) {
	t.Run("should log", func(t *testing.T) {
		expectedebugLogMessage := "a dbg test"
		expectedInfoLogMessage := "test message log"
		expectedWarnLogMessage := "a warn message"
		expectedErrLogMessage := "test err message log"

		// capture stdout
		saveStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		defer func() {
			os.Stdout = saveStdout
			_ = w.Close()
			_ = r.Close()
		}()

		logger := newSlogLogger(Config{Env: ProdEnvironment, LogLevel: DebugLevel})
		require.NotNil(t, logger)

		logger.Debug(expectedebugLogMessage)
		logger.Info(expectedInfoLogMessage)
		logger.Warn(expectedWarnLogMessage)
		logger.Error(expectedErrLogMessage)

		withLogger := logger.With("app", "myapp").WithContext(context.Background())
		withLogger.Info("another test")
		logger.Sync()

		_ = w.Close()
		os.Stdout = saveStdout

		// read captured output
		var buff bytes.Buffer
		_, err = io.Copy(&buff, r)
		require.NoError(t, err)
		_ = r.Close()

		output := buff.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, expectedebugLogMessage)
		assert.Contains(t, output, strings.ToUpper(DebugLevel.String()))
		assert.Contains(t, output, expectedInfoLogMessage)
		assert.Contains(t, output, strings.ToUpper(InfoLevel.String()))
		assert.Contains(t, output, expectedWarnLogMessage)
		assert.Contains(t, output, strings.ToUpper(WarnLevel.String()))
		assert.Contains(t, output, expectedErrLogMessage)
		assert.Contains(t, output, strings.ToUpper(ErrorLevel.String()))
		assert.Contains(t, output, "another test\",\"app\":\"myapp\"")
		assert.NotContains(t, output, "stack")
	})

	t.Run("should contains stack in dev env", func(t *testing.T) {
		expectedErrLogMessage := "test err message log"

		// capture stdout
		saveStdout := os.Stdout
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdout = w

		defer func() {
			os.Stdout = saveStdout
			_ = w.Close()
			_ = r.Close()
		}()

		logger := newSlogLogger(Config{Env: DevEnvironment, LogLevel: DebugLevel})
		require.NotNil(t, logger)

		logger.Error(expectedErrLogMessage)

		logger.Sync()

		_ = w.Close()
		os.Stdout = saveStdout

		// read captured output
		var buff bytes.Buffer
		_, err = io.Copy(&buff, r)
		require.NoError(t, err)
		_ = r.Close()

		output := buff.String()
		assert.NotEmpty(t, output)
		assert.Contains(t, output, expectedErrLogMessage)
		assert.Contains(t, output, strings.ToUpper(ErrorLevel.String()))
		assert.Contains(t, output, "stack")
	})
}

func TestLogLevel_Slog(t *testing.T) {
	cfg := Config{LogLevel: WarnLevel}
	logger := newSlogLogger(cfg)

	require.NotNil(t, logger)
	assert.Equal(t, strings.ToUpper(WarnLevel.String()), logger.LogLevel())
}

func TestHttpLevelHandler_Slog(t *testing.T) {
	t.Run("should change log level", func(t *testing.T) {
		body := strings.NewReader(`{"level":"debug"}`)
		req, err := http.NewRequest("PUT", "/loglevel", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		logger := newSlogLogger(Config{LogLevel: InfoLevel})
		require.NotNil(t, logger)

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return true })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, strings.ToUpper(DebugLevel.String()), logger.LogLevel())
	})

	t.Run("should return forbidden when auth handler return false", func(t *testing.T) {
		body := strings.NewReader(`{"level":"debug"}`)
		req, err := http.NewRequest("PUT", "/loglevel", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		logger := newSlogLogger(Config{LogLevel: InfoLevel})
		require.NotNil(t, logger)

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return false })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusForbidden)
		assert.Equal(t, strings.ToUpper(InfoLevel.String()), logger.LogLevel())
	})

	t.Run("should handle bad request for unknown log level", func(t *testing.T) {
		body := strings.NewReader(`{"level":"foo"}`)
		req, err := http.NewRequest("PUT", "/loglevel", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		logger := newSlogLogger(Config{LogLevel: InfoLevel})
		require.NotNil(t, logger)

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return true })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, strings.ToUpper(InfoLevel.String()), logger.LogLevel())
	})

	t.Run("should handle invalid payload", func(t *testing.T) {
		body := strings.NewReader(`{"foo"}`)
		req, err := http.NewRequest("PUT", "/loglevel", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		logger := newSlogLogger(Config{LogLevel: InfoLevel})
		require.NotNil(t, logger)

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return true })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusBadRequest)
		assert.Equal(t, strings.ToUpper(InfoLevel.String()), logger.LogLevel())
	})

	t.Run("should return current log level", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/loglevel", nil)
		require.NoError(t, err)
		rec := httptest.NewRecorder()

		logger := newSlogLogger(Config{LogLevel: InfoLevel})
		require.NotNil(t, logger)

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return true })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, strings.ToUpper(InfoLevel.String()), logger.LogLevel())
	})

	t.Run("should return mmethod not allowed", func(t *testing.T) {
		req, err := http.NewRequest("POST", "/loglevel", nil)
		require.NoError(t, err)
		rec := httptest.NewRecorder()

		logger := newSlogLogger(Config{LogLevel: InfoLevel})

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return true })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusMethodNotAllowed)
		assert.Equal(t, strings.ToUpper(InfoLevel.String()), logger.LogLevel())
	})
}
