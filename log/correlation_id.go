package log

import (
	"context"

	"github.com/google/uuid"
)

type correlationIDContextKey struct{}

func WithCorrelationID(c context.Context, correlationID string) context.Context {
	return context.WithValue(c, correlationIDContextKey{}, correlationID)
}

func CorrelationIDFromContext(c context.Context) string {
	if c == nil {
		return ""
	}
	id, ok := c.Value(correlationIDContextKey{}).(string)
	if !ok {
		return ""
	}
	return id
}

func NewCorrelationID() string {
	return uuid.New().String()
}
