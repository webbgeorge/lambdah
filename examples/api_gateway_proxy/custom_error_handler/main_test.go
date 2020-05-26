package main

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler_UnhandledError(t *testing.T) {
	h := newHandler().ToLambdaHandler()

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Headers: map[string]string{"Custom-Unhandled-Error": "true"},
	})

	assert.Nil(t, err)
	assert.Equal(t, `{"error":"internal_server_error","message":"Something went wrong"}`, res.Body)
}

func TestNewHandler_HandledError(t *testing.T) {
	h := newHandler().ToLambdaHandler()

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Headers: map[string]string{"Custom-Handled-Error": "true"},
	})

	assert.Nil(t, err)
	assert.Equal(t, `{"error":"my_custom_error","message":"Custom error triggered"}`, res.Body)
}
