package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler_Success_SkipOtherEvents(t *testing.T) {
	mockLogger := &bytes.Buffer{}
	h := newHandler(mockLogger).ToLambdaHandler()

	err := h(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{{EventName: "s3:test:event"}}},
	)

	assert.Nil(t, err)
	assert.Contains(t, mockLogger.String(), "Event name: 's3:test:event'")
}
