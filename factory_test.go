package azalogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	t.Run("should create zap logger based on backend", func(t *testing.T) {
		cfg := Config{
			Backend: ZapBackend,
		}

		logger, err := NewLogger(cfg)
		require.NoError(t, err)
		require.NotNil(t, logger)
		assert.IsType(t, &zapLogger{}, logger)
	})

	t.Run("should create zap logger based on backend", func(t *testing.T) {
		cfg := Config{
			Backend: SlogBackend,
		}

		logger, err := NewLogger(cfg)
		require.NoError(t, err)
		require.NotNil(t, logger)
		assert.IsType(t, &slogLogger{}, logger)
	})

	t.Run("should create in-memory logger based on backend", func(t *testing.T) {
		cfg := Config{
			Backend: InMemoryBackend,
		}

		logger, _ := NewLogger(cfg)
		require.NotNil(t, logger)
		assert.IsType(t, &InMemoryLogger{}, logger)
	})

	t.Run("should return an error on unsupported backend", func(t *testing.T) {
		cfg := Config{
			Backend: 1000,
		}

		logger, err := NewLogger(cfg)
		require.Error(t, err)
		assert.EqualError(t, err, ErrUnsupportedBackend.Error())
		assert.Nil(t, logger)
	})
}
