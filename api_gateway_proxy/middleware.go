package api_gateway_proxy

import (
	"fmt"
	"io"
	"net/http"

	"github.com/webbgeorge/lambdah/log"
)

type Middleware func(h HandlerFunc) HandlerFunc

// Middleware to handle errors returned by handler
//
// If returned error is of type Error{}, then a custom JSON response is
// returned in the response, with status Error{}.StatusCode and body
// `{"message": "Error{}.Message"}`
//
// If the logger middleware is also in use, this middleware will log a
// message for any errors.
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
				logger := log.LoggerFromContext(c.Context)
				if logger != nil {
					logger.Error().
						Str("error", err.Error()).
						Msgf("Error: %s", apiGatewayErr.Error())
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
// requests through distributed systems. The correlation ID is also returned in a
// response header so it can be used by consumers. You can also access the
// correlation ID directly in your handlers and middlewares by calling
// log.CorrelationIDFromContext(c.Context)
//
// First looks for a Correlation ID provided in the request header `Correlation-Id`.
// If not present a new correlation ID will be created and put into the request header.
//
// To ensure logs have correlation ID field, this middleware should be called first.
func CorrelationIDMiddleware() Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			cid := c.Request.Headers["Correlation-Id"]
			if cid == "" {
				cid = log.NewCorrelationID()
			}

			if c.Response.Headers == nil {
				c.Response.Headers = make(map[string]string)
			}
			c.Response.Headers["Correlation-Id"] = cid

			c.Context = log.WithCorrelationID(c.Context, cid)

			return h(c)
		}
	}
}

// Middleware to configure a logger in the context.Context, which is then used by other
// middleware (if available), including the error handler.
//
// The middleware logs on the response of each request, and also will include an
// error log message if the error handler middleware is also used. In addition to
// the default log messages, you can access the logger in your handlers/middleware
// by calling log.LoggerFromContext(c.Context). The logger we use is
// github.com/rs/zerolog
//
// w io.Writer              is the log output, for example os.Stdout
// field map[string]string  is a list of key value fields to include in each log message
//
// To ensure that the response status code is correctly reported, this middleware
// should be called before the error handler middleware.
func LoggerMiddleware(w io.Writer, fields map[string]string) Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			fields["handler_type"] = "api_gateway_proxy"
			fields["correlation_id"] = log.CorrelationIDFromContext(c.Context)
			fields["req_method"] = c.Request.HTTPMethod
			fields["req_path"] = c.Request.Path
			fields["req_route"] = c.Request.RequestContext.ResourcePath

			logger := log.NewLogger(w, fields)
			c.Context = log.WithLogger(c.Context, logger)
			err := h(c)

			logger.Info().
				Int("res_status", c.Response.StatusCode).
				Msgf("Response with status code %d", c.Response.StatusCode)

			return err
		}
	}
}
