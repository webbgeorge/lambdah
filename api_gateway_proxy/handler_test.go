package api_gateway_proxy

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/steinfletcher/apitest"
	jsonpath "github.com/steinfletcher/apitest-jsonpath"
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

	awsHandler := HandlerFunc(h).ToLambdaHandler()
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

func TestAPIGatewayProxyHandler_NoErrorHandler(t *testing.T) {
	h := func(c *Context) error {
		return errors.New("an error happened")
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	res, err := awsHandler(
		context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, 500, res.StatusCode)
	assert.Equal(t, `Internal server error`, res.Body)
}

func TestAPIGatewayProxyHandler_DefaultErrorHandler(t *testing.T) {
	h := func(c *Context) error {
		return errors.New("an error happened")
	}

	awsHandler := HandlerFunc(h).Middleware(ErrorHandlerMiddleware()).ToLambdaHandler()
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

	customErrorHandlerMiddleware := func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			err := h(c)
			if err != nil {
				type customError struct {
					Error string `json:"error"`
				}
				_ = c.JSON(http.StatusBadRequest, customError{Error: err.Error()})
			}
			return nil
		}
	}

	awsHandler := HandlerFunc(h).Middleware(customErrorHandlerMiddleware).ToLambdaHandler()

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

	awsHandler := HandlerFunc(h).Middleware(mw1, mw2).ToLambdaHandler()
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

// test HTTP handler function using steinfletcher/apitest
func TestHandlerFunc_ToHttpHandler(t *testing.T) {
	h := func(c *Context) error {
		var data struct {
			Age int
		}
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		assert.Equal(t, 12, data.Age)
		assert.Equal(t, "dog", c.Request.PathParameters["name"])
		assert.Len(t, c.Request.MultiValueHeaders["Multi-Header"], 2)
		assert.Equal(t, "val1", c.Request.Headers["Multi-Header"])

		c.Response.MultiValueHeaders = map[string][]string{
			"Test-Header": {"valueOne", "valueTwo"},
		}
		return c.JSON(http.StatusOK, responseData{Status: "all good"})
	}

	httpHandler := HandlerFunc(h).
		ToHttpHandler("/animal/{name}", nil)

	apitest.New().
		Handler(httpHandler).
		Put("/animal/dog").
		Header("Multi-Header", "val1").
		Header("Multi-Header", "val2").
		Body(`{"age": 12}`).
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Equal(`$.status`, "all good")).
		HeaderPresent("Test-Header").
		End()
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
