package dynamodb

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestDynamoDBHandler_Success_OneEvent(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		assert.Equal(t, "dynamodb:test:event", c.EventRecord.EventName)
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
			EventName: "dynamodb:test:event",
		}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, 1, callCount)
}

func TestDynamoDBHandler_Success_MultipleEvents(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		switch callCount {
		case 1:
			assert.Equal(t, "dynamodb:test:event:1", c.EventRecord.EventName)
		case 2:
			assert.Equal(t, "dynamodb:test:event:2", c.EventRecord.EventName)
		}
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{
			{
				EventName: "dynamodb:test:event:1",
			},
			{
				EventName: "dynamodb:test:event:2",
			},
		}},
	)

	assert.Nil(t, err)
	assert.Equal(t, 2, callCount)
}

func TestDynamoDBHandler_Error(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		return assert.AnError
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
			EventName: "dynamodb:test:event",
		}}},
	)

	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, 1, callCount)
}

func TestContext_UnmarshalKeys(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		EventRecord: events.DynamoDBEventRecord{
			EventName: "dynamodb:test:event",
			Change: events.DynamoDBStreamRecord{
				Keys: map[string]events.DynamoDBAttributeValue{
					"name": events.NewStringAttribute("Dave"),
					"age":  events.NewNumberAttribute("42"),
				},
			},
		},
	}

	var d testData
	err := c.BindKeys(&d)

	assert.Nil(t, err)
	assert.Equal(t, testData{
		Name: "Dave",
		Age:  42,
	}, d)
}

func TestContext_UnmarshalNewImage(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		EventRecord: events.DynamoDBEventRecord{
			EventName: "dynamodb:test:event",
			Change: events.DynamoDBStreamRecord{
				NewImage: map[string]events.DynamoDBAttributeValue{
					"name": events.NewStringAttribute("Dave"),
					"age":  events.NewNumberAttribute("42"),
				},
			},
		},
	}

	var d testData
	err := c.BindNewImage(&d)

	assert.Nil(t, err)
	assert.Equal(t, testData{
		Name: "Dave",
		Age:  42,
	}, d)
}

func TestContext_UnmarshalOldImage(t *testing.T) {
	c := &Context{
		Context: context.Background(),
		EventRecord: events.DynamoDBEventRecord{
			EventName: "dynamodb:test:event",
			Change: events.DynamoDBStreamRecord{
				OldImage: map[string]events.DynamoDBAttributeValue{
					"name": events.NewStringAttribute("Dave"),
					"age":  events.NewNumberAttribute("42"),
				},
			},
		},
	}

	var d testData
	err := c.BindOldImage(&d)

	assert.Nil(t, err)
	assert.Equal(t, testData{
		Name: "Dave",
		Age:  42,
	}, d)
}

func TestDynamoDBHandler_Middleware(t *testing.T) {
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
		events.DynamoDBEvent{Records: []events.DynamoDBEventRecord{{
			EventName: "dynamodb:test:event",
		}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, []string{"mw1 in", "mw2 in", "handler", "mw2 out", "mw1 out"}, callOrder)
}

type testData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
