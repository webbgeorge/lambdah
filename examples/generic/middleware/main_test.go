package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler_Success(t *testing.T) {
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockLogger).ToLambdaHandler()

	res, err := h(
		context.Background(),
		[]byte(`{"name":"Dave","greeting":"Hi"}`),
	)

	assert.Nil(t, err)
	assert.Equal(t, "ok", res)
	assert.Contains(t, mockLogger.String(), "name received: Dave")
}

func TestNewHandler_ParseError(t *testing.T) {
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockLogger).ToLambdaHandler()

	res, err := h(
		context.Background(),
		[]byte(`{"name`),
	)

	assert.NotNil(t, err)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
	assert.Nil(t, res)
	assert.Contains(t, mockLogger.String(), "Error processing generic event: unexpected end of JSON input")
}
