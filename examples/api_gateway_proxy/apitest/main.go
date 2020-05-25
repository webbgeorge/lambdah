package main

import (
	"errors"
	"net/http"

	"github.com/webbgeorge/lambdah/api_gateway_proxy"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	h := api_gateway_proxy.Handler(api_gateway_proxy.HandlerConfig{}, newHttpHandler())
	lambda.Start(h)
}

func newHttpHandler() api_gateway_proxy.HandlerFunc {
	return func(c *api_gateway_proxy.Context) error {
		if c.Request.Headers["Error"] == "true" {
			return errors.New("some error")
		}
		type responseData struct {
			Animal string `json:"animal"`
		}
		return c.JSON(http.StatusOK, responseData{Animal: "Giraffe"})
	}
}
