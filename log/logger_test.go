package log

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {
	c := WithLogger(context.Background(), NewLogger(ioutil.Discard, nil))
	assert.NotNil(t, LoggerFromContext(c))
}

func TestLoggerFromContext(t *testing.T) {
	// nil context
	assert.Nil(t, LoggerFromContext(nil))
	// without logger
	assert.Nil(t, LoggerFromContext(context.Background()))
	// with logger
	c := WithLogger(context.Background(), NewLogger(ioutil.Discard, nil))
	assert.NotNil(t, LoggerFromContext(c))
}

func TestNewLogger(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	logger := NewLogger(logBuffer, map[string]string{"field1": "value1"})
	logger.Info().Msg("Does it log?")
	assert.Contains(t, logBuffer.String(), "Does it log?")
}
