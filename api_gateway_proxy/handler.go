package api_gateway_proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// TODO: logging?

type HandlerFunc func(c *Context) error

type HandlerConfig struct {
	ErrorHandler ErrorHandler
}

func Handler(conf HandlerConfig, h HandlerFunc) func(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context,
		request events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		c := &Context{
			Context: ctx,
			Request: request,
		}

		err := h(c)
		if err != nil {
			if conf.ErrorHandler == nil {
				defaultErrorHandler(c, err)
			} else {
				conf.ErrorHandler(c, err)
			}
			return c.Response, nil
		}

		return c.Response, nil
	}
}

type Context struct {
	Context  context.Context
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

type Validatable interface {
	Validate() error
}

func (c *Context) Bind(v interface{}) error {
	err := json.Unmarshal([]byte(c.Request.Body), v)
	if err != nil {
		return err
	}

	if validatable, ok := v.(Validatable); ok {
		return validatable.Validate()
	}

	return nil
}

func (c *Context) JSON(statusCode int, body interface{}) error {
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		c.Response.Body = string(b)
	}
	c.Response.StatusCode = statusCode
	return nil
}

type ErrorHandler func(c *Context, err error)

type Error struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (err Error) Error() string {
	return fmt.Sprintf("status: %d, message: %s", err.StatusCode, err.Message)
}

func defaultErrorHandler(c *Context, err error) {
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
