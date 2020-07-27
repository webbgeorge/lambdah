package sns

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestSNSHandler_Success_OneEvent(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		assert.Equal(t, "sns:test:topic", c.EventRecord.SNS.TopicArn)
		assert.Equal(t, "test message", c.EventRecord.SNS.Message)
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.SNSEvent{Records: []events.SNSEventRecord{{
			SNS: events.SNSEntity{
				TopicArn: "sns:test:topic",
				Message:  "test message",
			},
		}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, 1, callCount)
}

func TestSNSHandler_Success_MultipleEvents(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		switch callCount {
		case 1:
			assert.Equal(t, "sns:test:topic:1", c.EventRecord.SNS.TopicArn)
			assert.Equal(t, "test message 1", c.EventRecord.SNS.Message)
		case 2:
			assert.Equal(t, "sns:test:topic:2", c.EventRecord.SNS.TopicArn)
			assert.Equal(t, "test message 2", c.EventRecord.SNS.Message)
		}
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.SNSEvent{Records: []events.SNSEventRecord{
			{
				SNS: events.SNSEntity{
					TopicArn: "sns:test:topic:1",
					Message:  "test message 1",
				},
			},
			{
				SNS: events.SNSEntity{
					TopicArn: "sns:test:topic:2",
					Message:  "test message 2",
				},
			},
		}},
	)

	assert.Nil(t, err)
	assert.Equal(t, 2, callCount)
}

func TestSNSHandler_Error(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		return assert.AnError
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.SNSEvent{Records: []events.SNSEventRecord{{
			SNS: events.SNSEntity{
				TopicArn: "sns:test:topic",
				Message:  "test message",
			},
		}}},
	)

	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, 1, callCount)
}

func TestSNSContext_Bind_Success(t *testing.T) {
	c := &Context{
		EventRecord: events.SNSEventRecord{
			SNS: events.SNSEntity{
				Message: `{"message": "hello"}`,
			},
		},
	}

	var data messageData
	err := c.Bind(&data)

	assert.Nil(t, err)
	assert.Equal(t, "hello", data.Message)
}

func TestSNSContext_Bind_InvalidJSON(t *testing.T) {
	c := &Context{
		EventRecord: events.SNSEventRecord{
			SNS: events.SNSEntity{
				Message: `{"messag`,
			},
		},
	}

	var data messageData
	err := c.Bind(&data)

	assert.Error(t, err)
}

func TestSNSHandler_Middleware(t *testing.T) {
	callOrder := make([]string, 0)

	h := func(c *Context) error {
		callOrder = append(callOrder, "handler")
		return nil
	}

	mw1 := func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			callOrder = append(callOrder, "mw1 in")
			err := h(c)
			callOrder = append(callOrder, "mw1 out")
			return err
		}
	}

	mw2 := func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			callOrder = append(callOrder, "mw2 in")
			err := h(c)
			callOrder = append(callOrder, "mw2 out")
			return err
		}
	}

	awsHandler := HandlerFunc(h).Middleware(mw1, mw2).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.SNSEvent{Records: []events.SNSEventRecord{{
			SNS: events.SNSEntity{
				TopicArn: "sns:test:topic",
				Message:  "test message",
			},
		}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, []string{"mw1 in", "mw2 in", "handler", "mw2 out", "mw1 out"}, callOrder)
}

type messageData struct {
	Message string `json:"message"`
}

func (d *messageData) Validate() error {
	if d.Message == "" {
		return errors.New("invalid message")
	}
	return nil
}
