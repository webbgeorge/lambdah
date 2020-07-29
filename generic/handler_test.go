package generic

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenericHandler_Success_OneEvent(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		assert.Equal(t, "test event", string(c.Event))
		c.Response = "ok"
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	res, err := awsHandler(
		context.Background(),
		[]byte("test event"),
	)

	assert.Nil(t, err)
	assert.Equal(t, "ok", res.(string))
	assert.Equal(t, 1, callCount)
}

func TestGenericHandler_Error(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		return assert.AnError
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	_, err := awsHandler(
		context.Background(),
		[]byte("test event"),
	)

	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, 1, callCount)
}

func TestGenericContext_Bind_Success(t *testing.T) {
	c := &Context{Event: []byte(`{"name":"Dave","age":45}`)}

	var data eventDetail
	err := c.Bind(&data)

	assert.Nil(t, err)
	assert.Equal(t, "Dave", data.Name)
}

func TestGenericContext_Bind_ValidationError(t *testing.T) {
	c := &Context{Event: []byte(`{"name":"Dave","age":"not an int"}`)}

	var data eventDetail
	err := c.Bind(&data)

	assert.NotNil(t, err)
	assert.Equal(t, "json: cannot unmarshal string into Go struct field eventDetail.age of type int", err.Error())
}

func TestGenericContext_Bind_InvalidJSON(t *testing.T) {
	c := &Context{Event: []byte(`{"name`)}

	var data eventDetail
	err := c.Bind(&data)

	assert.Error(t, err)
}

func TestGenericHandler_Middleware(t *testing.T) {
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
	_, err := awsHandler(
		context.Background(),
		[]byte("test event"),
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
