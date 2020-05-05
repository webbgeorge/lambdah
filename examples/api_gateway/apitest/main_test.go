package main

import (
	"net/http"
	"testing"

	"github.com/webbgeorge/lambdah"

	"github.com/steinfletcher/apitest"
	"github.com/steinfletcher/apitest-jsonpath"
)

func TestNewHandler_Success(t *testing.T) {
	h := lambdah.APIGatewayProxyHandler(lambdah.APIGatewayProxyHandlerConfig{}, newHttpHandler())

	apitest.New().
		Handler(lambdah.HttpHandlerFromAWSAPIGatewayProxyHandler(h, "/animal", nil)).
		Get("/animal").
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Equal(`$.animal`, "Giraffe")).
		End()
}

func TestNewHandler_WithError(t *testing.T) {
	h := lambdah.APIGatewayProxyHandler(lambdah.APIGatewayProxyHandlerConfig{}, newHttpHandler())

	apitest.New().
		Handler(lambdah.HttpHandlerFromAWSAPIGatewayProxyHandler(h, "/animal", nil)).
		Get("/animal").
		Header("Error", "true").
		Expect(t).
		Status(http.StatusInternalServerError).
		Assert(jsonpath.Equal(`$.message`, "Internal server error")).
		End()
}
