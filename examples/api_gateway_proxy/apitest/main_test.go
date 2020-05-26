package main

import (
	"net/http"
	"testing"

	"github.com/webbgeorge/lambdah/api_gateway_proxy"

	"github.com/steinfletcher/apitest"
	"github.com/steinfletcher/apitest-jsonpath"
)

func TestNewHandler_Success(t *testing.T) {
	httpHandler := api_gateway_proxy.ToHttpHandler(
		api_gateway_proxy.HandlerConfig{},
		newHttpHandler(),
		nil,
		"/animal",
		nil,
	)

	apitest.New().
		Handler(httpHandler).
		Get("/animal").
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Equal(`$.animal`, "Giraffe")).
		End()
}

func TestNewHandler_WithError(t *testing.T) {
	httpHandler := api_gateway_proxy.ToHttpHandler(
		api_gateway_proxy.HandlerConfig{},
		newHttpHandler(),
		nil,
		"/animal",
		nil,
	)

	apitest.New().
		Handler(httpHandler).
		Get("/animal").
		Header("Error", "true").
		Expect(t).
		Status(http.StatusInternalServerError).
		Assert(jsonpath.Equal(`$.message`, "Internal server error")).
		End()
}
