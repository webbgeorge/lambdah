package main

import (
	"context"
	"testing"

	"github.com/webbgeorge/lambdah"

	"github.com/aws/aws-lambda-go/events"
)

func TestNewHandler_Success(t *testing.T) {
	h := lambdah.APIGatewayProxyHandler(lambdah.APIGatewayProxyHandlerConfig{}, newHandler())

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Body: `{
				"greeting": "Hi",
				"name": "George"
			}`,
	})

	if err != nil {
		t.Log("unexpected error")
		t.Fail()
		return
	}

	expectedResponse := `{"message":"Hi George"}`

	if res.Body != expectedResponse {
		t.Logf("unexpected response '%s', '%s'", res.Body, expectedResponse)
		t.Fail()
	}
}

func TestNewHandler_MissingName(t *testing.T) {
	h := lambdah.APIGatewayProxyHandler(lambdah.APIGatewayProxyHandlerConfig{}, newHandler())

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Body: `{
				"greeting": "Hi"
			}`,
	})

	if err != nil {
		t.Log("unexpected error")
		t.Fail()
		return
	}

	expectedResponse := `{"message":"Name is required"}`

	if res.Body != expectedResponse {
		t.Logf("unexpected response '%s', '%s'", res.Body, expectedResponse)
		t.Fail()
	}
}

func TestNewHandler_InvalidGreeting(t *testing.T) {
	h := lambdah.APIGatewayProxyHandler(lambdah.APIGatewayProxyHandlerConfig{}, newHandler())

	res, err := h(context.Background(), events.APIGatewayProxyRequest{
		Body: `{
				"greeting": "Hey"
			}`,
	})

	if err != nil {
		t.Log("unexpected error")
		t.Fail()
		return
	}

	expectedResponse := `{"message":"Greeting not allowed"}`

	if res.Body != expectedResponse {
		t.Logf("unexpected response '%s', '%s'", res.Body, expectedResponse)
		t.Fail()
	}
}
