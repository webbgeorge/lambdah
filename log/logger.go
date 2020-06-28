package log

import (
	"context"
	"io"

	"github.com/rs/zerolog"
)

type logContextKey struct{}

func WithLogger(c context.Context, logger *zerolog.Logger) context.Context {
	return context.WithValue(c, logContextKey{}, logger)
}

func LoggerFromContext(c context.Context) *zerolog.Logger {
	if c == nil {
		return nil
	}
	logger, ok := c.Value(logContextKey{}).(*zerolog.Logger)
	if !ok {
		return nil
	}
	return logger
}

func NewLogger(w io.Writer, fields map[string]string) *zerolog.Logger {
	lc := zerolog.New(w).With().Timestamp()
	for field, value := range fields {
		lc = lc.Str(field, value)
	}
	logger := lc.Logger()
	return &logger
}
