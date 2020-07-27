package sqs

import (
	"bytes"
	"context"
	"testing"

	"github.com/webbgeorge/lambdah/log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/stretchr/testify/assert"
)

func TestCorrelationIDMiddleware_NoIdProvided(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		Message: events.SQSMessage{Body: "my message"},
	}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		assert.Len(t, log.CorrelationIDFromContext(c.Context), 36)
		return nil
	}

	mw := CorrelationIDMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
}

func TestCorrelationIDMiddleware_IdProvided(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		Message: events.SQSMessage{
			Body: "my message",
			MessageAttributes: map[string]events.SQSMessageAttribute{
				"correlation_id": {StringValue: aws.String("test-correlation-id")},
			},
		},
	}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		assert.Equal(t, "test-correlation-id", log.CorrelationIDFromContext(c.Context))
		return nil
	}

	mw := CorrelationIDMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
}

func TestLoggerMiddleware_Success(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		Message: events.SQSMessage{Body: "my message", MessageId: "message-1"},
	}
	h := func(c *Context) error {
		log.LoggerFromContext(c.Context).Info().Msg("msg from handler")
		return nil
	}
	buf := &bytes.Buffer{}

	mw := LoggerMiddleware(buf, map[string]string{"testField": "testFieldData"})
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.Contains(t, buf.String(), `"testField":"testFieldData"`)
	assert.Contains(t, buf.String(), "Processing SQS message, message ID 'message-1'")
	assert.Contains(t, buf.String(), "msg from handler")
}

func TestLoggerMiddleware_Error(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		Message: events.SQSMessage{Body: "my message", MessageId: "message-1"},
	}
	h := func(c *Context) error {
		log.LoggerFromContext(c.Context).Info().Msg("msg from handler")
		return assert.AnError
	}
	buf := &bytes.Buffer{}

	mw := LoggerMiddleware(buf, map[string]string{})
	h = mw(h)
	err := h(c)

	assert.NotNil(t, err)
	assert.Contains(t, buf.String(), "Processing SQS message, message ID 'message-1'")
	assert.Contains(t, buf.String(), "msg from handler")
	assert.Contains(t, buf.String(), "Error processing SQS message: assert.AnError general error for testing")
}
