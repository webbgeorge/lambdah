package main

import (
	"context"
	"testing"

	"github.com/webbgeorge/lambdah/api_gateway_proxy"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler_UnhandledError(t *testing.T) {
	h := api_gateway_proxy.Handler(
		api_gateway_proxy.HandlerConfig{
			ErrorHandler: customErrorHandler,
		},
		newHttpHandler(),
	)

	res, err := h(context.Background(), events.APIGatewayProxyRequest{})

	assert.Nil(t, err)
	assert.Equal(t, `{"error":"Something went wrong"}`, res.Body)
}

func TestNewHandler_CustomErrorType(t *testing.T) {
	h := api_gateway_proxy.Handler(
		api_gateway_proxy.HandlerConfig{
			ErrorHandler: customErrorHandler,
		},
		newHttpHandler(),
	)

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Headers: map[string]string{"Custom-Error": "true"},
	})

	assert.Nil(t, err)
	assert.Equal(t, `{"error":"Custom error triggered"}`, res.Body)
}
