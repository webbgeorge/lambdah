package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler_Success(t *testing.T) {
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockLogger).ToLambdaHandler()

	err := h(
		context.Background(),
		events.SQSEvent{Records: []events.SQSMessage{{
			Body: `{"name":"Dave","greeting":"Hi"}`,
		}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, "Dave", mockLogger.String())
}

func TestNewHandler_ValidationError(t *testing.T) {
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockLogger).ToLambdaHandler()

	err := h(
		context.Background(),
		events.SQSEvent{Records: []events.SQSMessage{{
			Body: `{"name":"Dave","greeting":"Hey"}`,
		}}},
	)

	assert.NotNil(t, err)
	assert.Equal(t, "greeting not allowed", err.Error())
	assert.Equal(t, "", mockLogger.String())
}

func TestNewHandler_ParseError(t *testing.T) {
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockLogger).ToLambdaHandler()

	err := h(
		context.Background(),
		events.SQSEvent{Records: []events.SQSMessage{{
			Body: `{"name`,
		}}},
	)

	assert.NotNil(t, err)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
	assert.Equal(t, "", mockLogger.String())
}
