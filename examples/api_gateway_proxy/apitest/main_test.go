package main

import (
	"net/http"
	"testing"

	"github.com/steinfletcher/apitest"
	"github.com/steinfletcher/apitest-jsonpath"
)

func TestNewHandler_Success(t *testing.T) {
	httpHandler := newHandler().
		ToHttpHandler("/animal/{name}", nil)

	apitest.New().
		Handler(httpHandler).
		Get("/animal/giraffe").
		Expect(t).
		Status(http.StatusOK).
		Assert(jsonpath.Equal(`$.animal`, "Giraffe")).
		Assert(jsonpath.Equal(`$.trait`, "tall")).
		End()
}

func TestNewHandler_UnhandledAnimalError(t *testing.T) {
	httpHandler := newHandler().
		ToHttpHandler("/animal/{name}", nil)

	apitest.New().
		Handler(httpHandler).
		Get("/animal/elephant").
		Expect(t).
		Status(http.StatusBadRequest).
		Assert(jsonpath.Equal(`$.message`, "Animal not found")).
		End()
}

func TestNewHandler_CatCausesUnhandledError(t *testing.T) {
	httpHandler := newHandler().
		ToHttpHandler("/animal/{name}", nil)

	apitest.New().
		Handler(httpHandler).
		Get("/animal/cat").
		Expect(t).
		Status(http.StatusInternalServerError).
		Assert(jsonpath.Equal(`$.message`, "Internal server error")).
		End()
}
