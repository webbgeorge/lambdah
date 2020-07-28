package dynamodb

import (
	"bytes"
	"context"
	"testing"

	"github.com/webbgeorge/lambdah/log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestCorrelationIDMiddleware(t *testing.T) {
	c := &Context{
		Context:     context.Background(),
		EventRecord: events.DynamoDBEventRecord{EventName: "dynamodb:test:event"},
	}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		assert.Equal(t, "dynamodb:test:event", c.EventRecord.EventName)
		assert.Len(t, log.CorrelationIDFromContext(c.Context), 36)
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
		Context:     context.Background(),
		EventRecord: events.DynamoDBEventRecord{EventName: "dynamodb:test:event"},
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
	assert.Contains(t, buf.String(), "Processing DynamoDB event 'dynamodb:test:event'")
	assert.Contains(t, buf.String(), "msg from handler")
}

func TestLoggerMiddleware_Error(t *testing.T) {
	c := &Context{
		Context:     context.Background(),
		EventRecord: events.DynamoDBEventRecord{EventName: "dynamodb:test:event"},
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
	assert.Contains(t, buf.String(), "Processing DynamoDB event 'dynamodb:test:event'")
	assert.Contains(t, buf.String(), "msg from handler")
	assert.Contains(t, buf.String(), "Error processing DynamoDB event: assert.AnError general error for testing")
}
