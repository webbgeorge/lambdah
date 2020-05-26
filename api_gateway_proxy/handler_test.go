package api_gateway_proxy

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestAPIGatewayProxyHandler_Success(t *testing.T) {
	h := func(c *Context) error {
		var data requestData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		assert.Equal(t, "hello", data.Message)

		return c.JSON(http.StatusOK, responseData{Status: "all good"})
	}

	awsHandler := Handler(HandlerConfig{}, h)
	res, err := awsHandler(
		context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, `{"status":"all good"}`, res.Body)
}

func TestAPIGatewayProxyHandler_DefaultErrorHandler(t *testing.T) {
	h := func(c *Context) error {
		return errors.New("an error happened")
	}

	awsHandler := Handler(HandlerConfig{}, h)
	res, err := awsHandler(
		context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, 500, res.StatusCode)
	assert.Equal(t, `{"message":"Internal server error"}`, res.Body)
}

func TestAPIGatewayProxyHandler_CustomErrorHandler(t *testing.T) {
	h := func(c *Context) error {
		return errors.New("an error happened")
	}

	awsHandler := Handler(HandlerConfig{
		ErrorHandler: func(c *Context, err error) {
			type customError struct {
				Error string `json:"error"`
			}
			_ = c.JSON(http.StatusBadRequest, customError{Error: err.Error()})
		},
	}, h)
	res, err := awsHandler(
		context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, `{"error":"an error happened"}`, res.Body)
}

func TestAPIGatewayProxyHandler_Middleware(t *testing.T) {
	callOrder := make([]string, 0)

	h := func(c *Context) error {
		callOrder = append(callOrder, "handler")
		return c.JSON(http.StatusOK, responseData{Status: c.Request.Headers["Test"]})
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

	awsHandler := Handler(HandlerConfig{}, h, mw1, mw2)
	res, err := awsHandler(
		context.Background(),
		events.APIGatewayProxyRequest{},
	)

	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, []string{"mw1 in", "mw2 in", "handler", "mw2 out", "mw1 out"}, callOrder)
}

func TestAPIGatewayProxyContext_Bind_Success(t *testing.T) {
	c := &Context{
		Request: events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	}

	var data requestData
	err := c.Bind(&data)

	assert.Nil(t, err)
	assert.Equal(t, "hello", data.Message)
}

func TestAPIGatewayProxyContext_Bind_InvalidJSON(t *testing.T) {
	c := &Context{
		Request: events.APIGatewayProxyRequest{
			Body: `{"messag`,
		},
	}

	var data requestData
	err := c.Bind(&data)

	assert.Error(t, err)
}

func TestAPIGatewayProxyContext_Bind_WithValidationError(t *testing.T) {
	c := &Context{
		Request: events.APIGatewayProxyRequest{
			Body: `{}`,
		},
	}

	var data requestData
	err := c.Bind(&data)

	assert.NotNil(t, err)
	assert.Equal(t, "invalid message", err.Error())
}

func TestAPIGatewayProxyContext_JSON_WithBody(t *testing.T) {
	c := &Context{}

	err := c.JSON(http.StatusOK, responseData{Status: "all good"})

	assert.Nil(t, err)
	assert.Equal(t, 200, c.Response.StatusCode)
	assert.Equal(t, `{"status":"all good"}`, c.Response.Body)
}

func TestAPIGatewayProxyContext_JSON_WithoutBody(t *testing.T) {
	c := &Context{}

	err := c.JSON(http.StatusNoContent, nil)

	assert.Nil(t, err)
	assert.Equal(t, 204, c.Response.StatusCode)
	assert.Equal(t, ``, c.Response.Body)
}

func TestDefaultAPIGatewayProxyErrorHandler_UnhandledError(t *testing.T) {
	c := &Context{}

	defaultErrorHandler(c, errors.New("some error"))

	assert.Equal(t, 500, c.Response.StatusCode)
	assert.Equal(t, `{"message":"Internal server error"}`, c.Response.Body)
}

func TestDefaultAPIGatewayProxyErrorHandler_APIGatewayProxyError(t *testing.T) {
	c := &Context{}

	defaultErrorHandler(c, Error{
		StatusCode: 400,
		Message:    "Bad request",
	})

	assert.Equal(t, 400, c.Response.StatusCode)
	assert.Equal(t, `{"message":"Bad request"}`, c.Response.Body)
}

type requestData struct {
	Message string `json:"message"`
}

func (d *requestData) Validate() error {
	if d.Message == "" {
		return errors.New("invalid message")
	}
	return nil
}

type responseData struct {
	Status string `json:"status"`
}
