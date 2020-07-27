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
		events.SNSEvent{Records: []events.SNSEventRecord{{
			SNS: events.SNSEntity{
				Message: `{"name":"Dave","greeting":"Hi"}`,
			},
		}}},
	)

	assert.Nil(t, err)
	assert.Contains(t, mockLogger.String(), "name received: Dave")
}

func TestNewHandler_ParseError(t *testing.T) {
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockLogger).ToLambdaHandler()

	err := h(
		context.Background(),
		events.SNSEvent{Records: []events.SNSEventRecord{{
			SNS: events.SNSEntity{
				Message: `{"name`,
			},
		}}},
	)

	assert.NotNil(t, err)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
	assert.Contains(t, mockLogger.String(), "Error processing SNS event: unexpected end of JSON input")
}
