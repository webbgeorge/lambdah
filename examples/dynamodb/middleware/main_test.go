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
		events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
			EventName: "dynamodb:test:event",
		}}},
	)

	assert.Nil(t, err)
	assert.Contains(t, mockLogger.String(), "Event name: 'dynamodb:test:event'")
}
