package main

import (
	"errors"
	"net/http"

	"github.com/webbgeorge/lambdah"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	h := lambdah.APIGatewayProxyHandler(
		lambdah.APIGatewayProxyHandlerConfig{
			ErrorHandler: customErrorHandler,
		},
		newHttpHandler(),
	)
	lambda.Start(h)
}

func newHttpHandler() lambdah.APIGatewayProxyHandlerFunc {
	return func(c *lambdah.APIGatewayProxyContext) error {
		if c.Request.Headers["Custom-Error"] == "true" {
			return customError{
				StatusCode:   400,
				ErrorMessage: "Custom error triggered",
			}
		}
		return errors.New("some error")
	}
}

type customError struct {
	StatusCode   int    `json:"-"`
	ErrorMessage string `json:"error"`
}

func (err customError) Error() string {
	return err.ErrorMessage
}

func customErrorHandler(c *lambdah.APIGatewayProxyContext, err error) {
	var customErr customError
	switch err := err.(type) {
	case customError:
		customErr = err
	default:
		customErr = customError{
			StatusCode:   http.StatusInternalServerError,
			ErrorMessage: "Something went wrong",
		}
	}
	_ = c.JSON(customErr.StatusCode, customErr)
}
