package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithCorrelationID(t *testing.T) {
	c := WithCorrelationID(context.Background(), "my-id")
	assert.Equal(t, "my-id", CorrelationIDFromContext(c))
}

func TestCorrelationIDFromContext(t *testing.T) {
	// nil context
	assert.Equal(t, "", CorrelationIDFromContext(nil))
	// without id
	assert.Equal(t, "", CorrelationIDFromContext(context.Background()))
	// with id
	c := WithCorrelationID(context.Background(), "my-id")
	assert.Equal(t, "my-id", CorrelationIDFromContext(c))
}

func TestNewCorrelationID(t *testing.T) {
	id := NewCorrelationID()
	assert.Len(t, id, 36)
}
