package api_gateway_proxy

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

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
		assert.NotEmpty(t, c.Request.Headers["Correlation-Id"])
		assert.NotEmpty(t, c.Request.MultiValueHeaders["Correlation-Id"][0])
		return nil
	}

	mw := CorrelationIDMiddleware()
	h = mw(h)
	err := h(c)

	assert.Nil(t, err)
	assert.True(t, handlerCalled)
}
