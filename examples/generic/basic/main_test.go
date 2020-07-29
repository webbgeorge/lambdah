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
	assert.Equal(t, "Dave", mockLogger.String())
}

func TestNewHandler_ValidationError(t *testing.T) {
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockLogger).ToLambdaHandler()

	res, err := h(
		context.Background(),
		[]byte(`{"name":"Dave","greeting":"Hey"}`),
	)

	assert.NotNil(t, err)
	assert.Equal(t, "greeting not allowed", err.Error())
	assert.Nil(t, res)
	assert.Equal(t, "", mockLogger.String())
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
	assert.Equal(t, "", mockLogger.String())
}
