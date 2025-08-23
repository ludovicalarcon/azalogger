package azalogger

import (
	"errors"
	"fmt"
)

var ErrUnsupportedBackend = errors.New("unsupported logger backend")

func NewLogger(cfg Config) (Logger, error) {
	switch cfg.Backend {
	case ZapBackend:
		return newZapLogger(cfg)
	case SlogBackend:
		return newSlogLogger(cfg), nil
	case InMemoryBackend:
		fmt.Println("calling NewMemoryLogger(cfg) directly is prefered")
		return NewInMemoryLogger(cfg), nil
	default:
		return nil, ErrUnsupportedBackend
	}
}
