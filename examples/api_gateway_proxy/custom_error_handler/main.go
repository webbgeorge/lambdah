package main

import (
	"errors"
	"net/http"

	"github.com/webbgeorge/lambdah/api_gateway_proxy"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	h := api_gateway_proxy.Handler(
		api_gateway_proxy.HandlerConfig{
			ErrorHandler: customErrorHandler,
		},
		newHttpHandler(),
	)
	lambda.Start(h)
}

func newHttpHandler() api_gateway_proxy.HandlerFunc {
	return func(c *api_gateway_proxy.Context) error {
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

func customErrorHandler(c *api_gateway_proxy.Context, err error) {
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
