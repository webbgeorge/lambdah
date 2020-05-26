package main

import (
	"errors"
	"fmt"
	"net/http"

	lambdah "github.com/webbgeorge/lambdah/api_gateway_proxy"
)

func main() {
	newHandler().Start()
}

func newHandler() lambdah.HandlerFunc {
	return lambdah.HandlerFunc(func(c *lambdah.Context) error {
		if c.Request.Headers["Custom-Handled-Error"] == "true" {
			return customError{
				StatusCode:   400,
				ErrorName:    "my_custom_error",
				ErrorMessage: "Custom error triggered",
			}
		} else if c.Request.Headers["Custom-Unhandled-Error"] == "true" {
			return errors.New("some error")
		}
		return nil
	}).Middleware(customErrorHandlerMiddleware())
}

type customError struct {
	StatusCode   int    `json:"-"`
	ErrorName    string `json:"error"`
	ErrorMessage string `json:"message"`
}

func (err customError) Error() string {
	return fmt.Sprintf("%s: %s", err.ErrorName, err.ErrorMessage)
}

func customErrorHandlerMiddleware() lambdah.Middleware {
	return func(h lambdah.HandlerFunc) lambdah.HandlerFunc {
		return func(c *lambdah.Context) error {
			err := h(c)
			if err != nil {
				var customErr customError
				switch err := err.(type) {
				case customError:
					customErr = err
				default:
					customErr = customError{
						StatusCode:   http.StatusInternalServerError,
						ErrorName:    "internal_server_error",
						ErrorMessage: "Something went wrong",
					}
				}
				_ = c.JSON(customErr.StatusCode, customErr)
			}
			return nil
		}
	}
}
