package main

import (
	"context"
	"testing"

	"github.com/webbgeorge/lambdah/api_gateway_proxy"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler_Success(t *testing.T) {
	h := api_gateway_proxy.Handler(api_gateway_proxy.HandlerConfig{}, newHandler())

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Body: `{
				"greeting": "Hi",
				"name": "George"
			}`,
	})

	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, `{"message":"Hi George"}`, res.Body)
}

func TestNewHandler_MissingName(t *testing.T) {
	h := api_gateway_proxy.Handler(api_gateway_proxy.HandlerConfig{}, newHandler())

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Body: `{
				"greeting": "Hi"
			}`,
	})

	assert.Nil(t, err)
	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, `{"message":"Name is required"}`, res.Body)
}

func TestNewHandler_InvalidGreeting(t *testing.T) {
	h := api_gateway_proxy.Handler(api_gateway_proxy.HandlerConfig{}, newHandler())

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Body: `{
				"greeting": "Hey"
			}`,
	})

	assert.Nil(t, err)
	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, `{"message":"Greeting not allowed"}`, res.Body)
}

func TestNewHandler_DaveIsNotWelcome(t *testing.T) {
	h := api_gateway_proxy.Handler(api_gateway_proxy.HandlerConfig{}, newHandler())

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Body: `{
				"greeting": "Hi",
				"name": "Dave"
			}`,
	})

	assert.Nil(t, err)
	assert.Equal(t, 406, res.StatusCode)
	assert.Equal(t, `{"message":"Dave is not welcome here!"}`, res.Body)
}
