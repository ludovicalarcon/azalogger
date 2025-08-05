package azalogger

import "errors"

var ErrUnsupportedBackend = errors.New("unsupported logger backend")

func NewLogger(cfg Config) (Logger, error) {
	switch cfg.Backend {
	case ZapBackend:
		return newZapLogger(cfg)
	case InMemoryBackend:
		return newInMemoryLogger(cfg), nil
	default:
		return nil, ErrUnsupportedBackend
	}
}
