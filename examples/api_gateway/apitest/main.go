package main

import (
	"errors"
	"net/http"

	"github.com/webbgeorge/lambdah"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	h := lambdah.APIGatewayProxyHandler(lambdah.APIGatewayProxyHandlerConfig{}, newHttpHandler())
	lambda.Start(h)
}

func newHttpHandler() lambdah.APIGatewayProxyHandlerFunc {
	return func(c *lambdah.APIGatewayProxyContext) error {
		if c.Request.Headers["Error"] == "true" {
			return errors.New("some error")
		}
		type responseData struct {
			Animal string `json:"animal"`
		}
		return c.JSON(http.StatusOK, responseData{Animal: "Giraffe"})
	}
}
