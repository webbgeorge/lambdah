package cloudwatch_events

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestCloudWatchEventsHandler_Success_OneEvent(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		assert.Equal(t, "test detail type", c.Event.DetailType)
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.CloudWatchEvent{DetailType: "test detail type"},
	)

	assert.Nil(t, err)
	assert.Equal(t, 1, callCount)
}

func TestCloudWatchEventsHandler_Error(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		return assert.AnError
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.CloudWatchEvent{DetailType: "test detail type"},
	)

	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, 1, callCount)
}

func TestCloudWatchEventsContext_Bind_Success(t *testing.T) {
	c := &Context{Event: events.CloudWatchEvent{
		DetailType: "test detail type",
		Detail:     []byte(`{"name":"Dave","age":35}`),
	}}

	var data eventDetail
	err := c.Bind(&data)

	assert.Nil(t, err)
	assert.Equal(t, "Dave", data.Name)
}

func TestCloudWatchEventsContext_Bind_ValidationError(t *testing.T) {
	c := &Context{Event: events.CloudWatchEvent{
		DetailType: "test detail type",
		Detail:     []byte(`{"name":"Dave","age":"not a number"}`),
	}}

	var data eventDetail
	err := c.Bind(&data)

	assert.NotNil(t, err)
	assert.Equal(t, "json: cannot unmarshal string into Go struct field eventDetail.age of type int", err.Error())
}

func TestCloudWatchEventsContext_Bind_InvalidJSON(t *testing.T) {
	c := &Context{Event: events.CloudWatchEvent{
		DetailType: "test detail type",
		Detail:     []byte(`{"na`),
	}}

	var data eventDetail
	err := c.Bind(&data)

	assert.Error(t, err)
}

func TestCloudWatchEventsHandler_Middleware(t *testing.T) {
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
		events.CloudWatchEvent{DetailType: "test detail type"},
	)

	assert.Nil(t, err)
	assert.Equal(t, []string{"mw1 in", "mw2 in", "handler", "mw2 out", "mw1 out"}, callOrder)
}

type eventDetail struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func (d *eventDetail) Validate() error {
	if d.Name == "" {
		return errors.New("invalid message")
	}
	if d.Age < 1 {
		return errors.New("invalid age")
	}
	return nil
}
