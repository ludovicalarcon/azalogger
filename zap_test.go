package azalogger

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCreateZapConfig(t *testing.T) {
	t.Run("should create zap config based on config (dev)", func(t *testing.T) {
		expected := zap.NewDevelopmentEncoderConfig()
		cfg := Config{
			LogLevel: WarnLevel,
			Env:      DevEnvironment,
		}

		got := createZapConfig(cfg)
		assert.Equal(t, WarnLevel.String(), got.Level.String())
		assert.Equal(t, "console", got.Encoding)
		assert.Equal(t, expected.MessageKey, got.EncoderConfig.MessageKey)
		assert.Equal(t, expected.LevelKey, got.EncoderConfig.LevelKey)
		assert.Equal(t, "timestamp", got.EncoderConfig.TimeKey)
		assert.Equal(t, expected.NameKey, got.EncoderConfig.NameKey)
		assert.Equal(t, expected.CallerKey, got.EncoderConfig.CallerKey)
		assert.Equal(t, expected.FunctionKey, got.EncoderConfig.FunctionKey)
		assert.Equal(t, expected.StacktraceKey, got.EncoderConfig.StacktraceKey)
	})

	t.Run("should create zap config based on config (prod)", func(t *testing.T) {
		expected := zap.NewProductionEncoderConfig()
		cfg := Config{
			LogLevel: WarnLevel,
			Env:      ProdEnvironment,
		}

		got := createZapConfig(cfg)
		assert.Equal(t, WarnLevel.String(), got.Level.String())
		assert.Equal(t, "json", got.Encoding)
		assert.Equal(t, expected.MessageKey, got.EncoderConfig.MessageKey)
		assert.Equal(t, expected.LevelKey, got.EncoderConfig.LevelKey)
		assert.Equal(t, "timestamp", got.EncoderConfig.TimeKey)
		assert.Equal(t, expected.NameKey, got.EncoderConfig.NameKey)
		assert.Equal(t, expected.CallerKey, got.EncoderConfig.CallerKey)
		assert.Equal(t, expected.FunctionKey, got.EncoderConfig.FunctionKey)
		assert.Equal(t, expected.StacktraceKey, got.EncoderConfig.StacktraceKey)
	})

	t.Run("should default to info loglevel when unknown one are passed", func(t *testing.T) {
		t.Setenv(LogLevelEnvVar, "unknown")
		got := createZapConfig(Config{})

		assert.Equal(t, InfoLevel.String(), got.Level.String())
		assert.Equal(t, "json", got.Encoding)
	})
}

func TestGetLogLevel(t *testing.T) {
	t.Run("should override config loglevel when env var is set", func(t *testing.T) {
		cfg := Config{
			LogLevel: InfoLevel,
		}
		expected := DebugLevel
		t.Setenv(LogLevelEnvVar, "debug")

		got := getLogLevel(cfg)

		assert.Equal(t, expected, got)
	})

	t.Run("should default to info when loglevel not set", func(t *testing.T) {
		got := getLogLevel(Config{})
		assert.Equal(t, InfoLevel, got)
	})

	t.Run("should use config when env var is not set", func(t *testing.T) {
		expected := WarnLevel
		cfg := Config{
			LogLevel: expected,
		}

		got := getLogLevel(cfg)

		assert.Equal(t, expected, got)
	})
}

func TestLogs(t *testing.T) {
	expectedebugLogMessage := "a dbg test"
	expectedInfoLogMessage := "test message log"
	expectedWarnLogMessage := "a warn message"
	expectedErrLogMessage := "test err message log"

	// capture stderr
	saveStdErr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	defer func() {
		os.Stderr = saveStdErr
		_ = w.Close()
		_ = r.Close()
	}()

	logger, err := newZapLogger(Config{Env: ProdEnvironment, LogLevel: DebugLevel})
	require.NoError(t, err)
	require.NotNil(t, logger)

	logger.Debug(expectedebugLogMessage)
	logger.Info(expectedInfoLogMessage)
	logger.Warn(expectedWarnLogMessage)
	logger.Error(expectedErrLogMessage)

	withLogger := logger.With("app", "myapp").WithContext(context.Background())
	withLogger.Info("another test")
	logger.Sync()

	_ = w.Close()
	os.Stderr = saveStdErr

	// read captured output
	var buff bytes.Buffer
	_, err = io.Copy(&buff, r)
	require.NoError(t, err)
	_ = r.Close()

	output := buff.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, expectedebugLogMessage)
	assert.Contains(t, output, DebugLevel.String())
	assert.Contains(t, output, expectedInfoLogMessage)
	assert.Contains(t, output, InfoLevel.String())
	assert.Contains(t, output, expectedWarnLogMessage)
	assert.Contains(t, output, WarnLevel.String())
	assert.Contains(t, output, expectedErrLogMessage)
	assert.Contains(t, output, ErrorLevel.String())
	assert.Contains(t, output, "another test\",\"app\":\"myapp\"")
}

func TestLogLevel_Zap(t *testing.T) {
	cfg := Config{LogLevel: WarnLevel}
	logger, err := newZapLogger(cfg)

	require.NoError(t, err)
	require.NotNil(t, logger)
	assert.Equal(t, WarnLevel.String(), logger.LogLevel())
}

func TestHttpLevelHandler_Zap(t *testing.T) {
	t.Run("should change log level", func(t *testing.T) {
		body := strings.NewReader(`{"level":"debug"}`)
		req, err := http.NewRequest("PUT", "/loglevel", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		logger, err := newZapLogger(Config{LogLevel: InfoLevel})
		require.NoError(t, err)

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return true })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, DebugLevel.String(), logger.LogLevel())
	})

	t.Run("should return forbidden when auth handler return false", func(t *testing.T) {
		body := strings.NewReader(`{"level":"debug"}`)
		req, err := http.NewRequest("PUT", "/loglevel", body)
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		logger, err := newZapLogger(Config{LogLevel: InfoLevel})
		require.NoError(t, err)

		handler := logger.HTTPLevelHandler(func(req *http.Request) bool { return false })
		handler.ServeHTTP(rec, req)

		assert.Equal(t, rec.Code, http.StatusForbidden)
		assert.Equal(t, InfoLevel.String(), logger.LogLevel())
	})
}
