package api_gateway_proxy

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/webbgeorge/lambdah/log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestErrorHandlerMiddleware_NoError(t *testing.T) {
	c := &Context{}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		return c.JSON(http.StatusOK, responseData{Status: "all good"})
	}

	mw := ErrorHandlerMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, c.Response.StatusCode, 200)
	assert.Equal(t, c.Response.Headers["Content-Type"], "application/json")
	assert.Equal(t, c.Response.Body, `{"status":"all good"}`)
}

func TestErrorHandlerMiddleware_UnhandledError(t *testing.T) {
	c := &Context{}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		return assert.AnError
	}

	mw := ErrorHandlerMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, c.Response.StatusCode, 500)
	assert.Equal(t, c.Response.Headers["Content-Type"], "application/json")
	assert.Equal(t, c.Response.Body, `{"message":"Internal server error"}`)
}

func TestErrorHandlerMiddleware_CustomError(t *testing.T) {
	c := &Context{}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		return Error{
			StatusCode: 400,
			Message:    "Bad request",
		}
	}

	mw := ErrorHandlerMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
	assert.Equal(t, c.Response.StatusCode, 400)
	assert.Equal(t, c.Response.Headers["Content-Type"], "application/json")
	assert.Equal(t, c.Response.Body, `{"message":"Bad request"}`)
}

func TestCorrelationIDMiddleware_CorrelationIDProvidedInRequest(t *testing.T) {
	c := &Context{
		Request: events.APIGatewayProxyRequest{
			Headers:           map[string]string{"Correlation-Id": "123abc"},
			MultiValueHeaders: map[string][]string{"Correlation-Id": {"123abc"}},
		},
	}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		assert.Equal(t, "123abc", c.Request.Headers["Correlation-Id"])
		assert.Equal(t, "123abc", c.Request.MultiValueHeaders["Correlation-Id"][0])
		return nil
	}

	mw := CorrelationIDMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
}

func TestCorrelationIDMiddleware_CorrelationIDNotProvidedInRequest(t *testing.T) {
	c := &Context{}
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		assert.NotEmpty(t, log.CorrelationIDFromContext(c.Context))
		return nil
	}

	mw := CorrelationIDMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
}

func TestLoggerMiddleware(t *testing.T) {
	c := &Context{Request: events.APIGatewayProxyRequest{
		HTTPMethod: "PUT",
		Path:       "/path/123",
		RequestContext: events.APIGatewayProxyRequestContext{
			ResourcePath: "/path/{pathID}",
		},
	}}
	c.Context = log.WithCorrelationID(c.Context, log.NewCorrelationID())
	handlerCalled := false
	h := func(c *Context) error {
		handlerCalled = true
		return c.JSON(http.StatusOK, nil)
	}
	buf := &bytes.Buffer{}

	mw := LoggerMiddleware(buf, map[string]string{"one": "val1", "two": "val2"})
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)

	var line map[string]interface{}
	err = json.Unmarshal(buf.Bytes(), &line)
	assert.Nil(t, err)
	assert.Equal(t, "info", line["level"])
	assert.Equal(t, "PUT", line["req_method"])
	assert.Equal(t, "/path/123", line["req_path"])
	assert.Equal(t, "/path/{pathID}", line["req_route"])
	assert.NotEmpty(t, line["correlation_id"])
	assert.Equal(t, float64(200), line["res_status"])
	assert.Equal(t, "Response with status code 200", line["message"])
}
