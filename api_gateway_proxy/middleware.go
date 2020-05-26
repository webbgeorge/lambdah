package api_gateway_proxy

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

type Middleware func(h HandlerFunc) HandlerFunc

// Middleware to handle errors returned by handler
//
// If returned error is of type Error{}, then a custom JSON response is
// returned in the response, with status Error{}.StatusCode and body
// `{"message": "Error{}.Message"}`
//
// If you wish to build a custom error handler you can use this pattern
// and create a middleware and an Error type with your own structure.
func ErrorHandlerMiddleware() Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			err := h(c)
			if err != nil {
				var apiGatewayErr Error
				switch err := err.(type) {
				case Error:
					apiGatewayErr = err
				default:
					apiGatewayErr = Error{
						StatusCode: http.StatusInternalServerError,
						Message:    "Internal server error",
					}
				}
				_ = c.JSON(apiGatewayErr.StatusCode, apiGatewayErr)
			}
			return nil
		}
	}
}

type Error struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (err Error) Error() string {
	return fmt.Sprintf("status: %d, message: %s", err.StatusCode, err.Message)
}

// Middleware to get or attach a correlation ID to the request, useful for tracing
// requests through distributed systems.
//
// First looks for a Correlation ID provided in the request header `Correlation-Id`.
// If not present a new correlation ID will be created and put into the request header.
func CorrelationIDMiddleware() Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			cid := c.Request.Headers["Correlation-Id"]
			if cid == "" {
				cid = uuid.New().String()

				if c.Request.Headers == nil {
					c.Request.Headers = make(map[string]string)
				}
				c.Request.Headers["Correlation-Id"] = cid

				if c.Request.MultiValueHeaders == nil {
					c.Request.MultiValueHeaders = make(map[string][]string)
				}
				c.Request.MultiValueHeaders["Correlation-Id"] = []string{cid}
			}
			return h(c)
		}
	}
}
