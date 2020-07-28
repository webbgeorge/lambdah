package main

import (
	"bytes"
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler_Success(t *testing.T) {
	logMock := &bytes.Buffer{}

	h := newHandler(logMock).ToLambdaHandler()

	err := h(
		context.Background(),
		events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
			EventName: "dynamodb:test:event",
			Change: events.DynamoDBStreamRecord{
				NewImage: map[string]events.DynamoDBAttributeValue{
					"name": events.NewStringAttribute("Dave"),
					"age":  events.NewNumberAttribute("42"),
				},
			},
		}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, "dynamodb change of type 'dynamodb:test:event' for name 'Dave'", logMock.String())
}

func TestNewHandler_Error(t *testing.T) {
	logMock := &bytes.Buffer{}

	h := newHandler(logMock).ToLambdaHandler()

	err := h(
		context.Background(),
		events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
			EventName: "dynamodb:test:event",
			Change: events.DynamoDBStreamRecord{
				NewImage: map[string]events.DynamoDBAttributeValue{
					// incorrect type
					"name": events.NewBooleanAttribute(false),
				},
			},
		}}},
	)

	assert.NotNil(t, err)
	assert.Empty(t, logMock.String())
}
