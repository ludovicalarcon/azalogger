package azalogger

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryLogger(t *testing.T) {
	t.Run("should create logger with proper loglevel", func(t *testing.T) {
		testCases := []struct {
			name     string
			cfg      Config
			expected string
		}{
			{
				name:     "debug",
				cfg:      Config{LogLevel: DebugLevel},
				expected: DebugLevel.String(),
			},
			{
				name:     "info",
				cfg:      Config{LogLevel: InfoLevel},
				expected: InfoLevel.String(),
			},
			{
				name:     "warn",
				cfg:      Config{LogLevel: WarnLevel},
				expected: WarnLevel.String(),
			},
			{
				name:     "error",
				cfg:      Config{LogLevel: ErrorLevel},
				expected: ErrorLevel.String(),
			},
			{
				name:     "fatal",
				cfg:      Config{LogLevel: FatalLevel},
				expected: FatalLevel.String(),
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logger := newInMemoryLogger(tc.cfg)
				assert.Equal(t, tc.expected, logger.logLevel.String())
			})
		}
	})

	t.Run("should default to info logLevel on invalid or empty", func(t *testing.T) {
		l1 := newInMemoryLogger(Config{})
		l2 := newInMemoryLogger(Config{LogLevel: "foo"})

		assert.Equal(t, InfoLevel.String(), l1.logLevel.String())
		assert.Equal(t, InfoLevel.String(), l2.logLevel.String())
	})
}

func TestInMemoryLogger(t *testing.T) {
	t.Run("should log", func(t *testing.T) {
		logger := newInMemoryLogger(Config{LogLevel: InfoLevel})
		expectedLog := "a test log"

		logger.Info(expectedLog)

		entries := logger.Entries()
		require.Len(t, entries, 2)
		assert.Equal(t, "[INFO] "+expectedLog, entries[0])
	})

	t.Run("should take loglevel into consideration", func(t *testing.T) {
		testCases := []struct {
			name     string
			logLevel LogLevel
			expected []string
		}{
			{
				name:     "debug",
				logLevel: DebugLevel,
				expected: []string{"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]", "[FATAL]"},
			},
			{
				name:     "info",
				logLevel: InfoLevel,
				expected: []string{"[INFO]", "[WARN]", "[ERROR]", "[FATAL]"},
			},
			{
				name:     "warn",
				logLevel: WarnLevel,
				expected: []string{"[WARN]", "[ERROR]", "[FATAL]"},
			},
			{
				name:     "error",
				logLevel: ErrorLevel,
				expected: []string{"[ERROR]", "[FATAL]"},
			},
			{
				name:     "fatal",
				logLevel: FatalLevel,
				expected: []string{"[FATAL]"},
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logger := newInMemoryLogger(Config{LogLevel: tc.logLevel})
				expectedLog := "test"

				logger.Debug(expectedLog, "app", "myapp")
				logger.Info(expectedLog, "app", "myapp")
				logger.Warn(expectedLog, "app", "myapp")
				logger.Error(expectedLog, "app", "myapp")
				logger.Fatal(expectedLog, "app", "myapp")
				logger.Sync()

				entries := logger.Entries()

				require.Len(t, entries, len(tc.expected)+1)
				for i, want := range tc.expected {
					assert.Equal(t, fmt.Sprintf("%s %s app=myapp", want, expectedLog), entries[i])
				}
			})
		}
	})

	t.Run("should inject fields", func(t *testing.T) {
		expectedLog := "some other log"

		withLogger := newInMemoryLogger(Config{}).With("foo", "bar").WithContext(context.Background())
		withLogger.Warn(expectedLog)

		entries := withLogger.(*InMemoryLogger).Entries()
		require.Len(t, entries, 2)
		assert.Equal(t, fmt.Sprintf("%s %s foo=bar", "[WARN]", expectedLog), entries[0])
	})

	t.Run("should ignore injected fields when it's not a kv pair", func(t *testing.T) {
		expectedLog := "another log"

		withLogger := newInMemoryLogger(Config{}).With("foo").WithContext(context.Background())
		withLogger.Warn(expectedLog)

		entries := withLogger.(*InMemoryLogger).Entries()
		require.Len(t, entries, 2)
		assert.Equal(t, fmt.Sprintf("%s %s", "[WARN]", expectedLog), entries[0])
	})
}

func TestLogLevel_InMemory(t *testing.T) {
	cfg := Config{LogLevel: WarnLevel}
	logger := newInMemoryLogger(cfg)

	assert.Equal(t, WarnLevel.String(), logger.LogLevel())
}

func TestHttpLevelHandler_InMemory(t *testing.T) {
	body := strings.NewReader(`{"level":"debug"}`)
	req, err := http.NewRequest("PUT", "/loglevel", body)
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	logger := newInMemoryLogger(Config{LogLevel: DebugLevel})

	handler := logger.HTTPLevelHandler()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, rec.Code, http.StatusNotImplemented)
}
