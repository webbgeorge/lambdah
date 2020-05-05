package lambdah

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// TODO: logging?

type APIGatewayProxyHandlerFunc func(c *APIGatewayProxyContext) error

type APIGatewayProxyHandlerConfig struct {
	ErrorHandler APIGatewayProxyErrorHandler
}

func APIGatewayProxyHandler(conf APIGatewayProxyHandlerConfig, h APIGatewayProxyHandlerFunc) func(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context,
		request events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		c := &APIGatewayProxyContext{
			Context: ctx,
			Request: request,
		}

		err := h(c)
		if err != nil {
			if conf.ErrorHandler == nil {
				defaultAPIGatewayProxyErrorHandler(c, err)
			} else {
				conf.ErrorHandler(c, err)
			}
			return c.Response, nil
		}

		return c.Response, nil
	}
}

type APIGatewayProxyContext struct {
	Context  context.Context
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

type Validatable interface {
	Validate() error
}

type APIGatewayProxyError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (err APIGatewayProxyError) Error() string {
	return fmt.Sprintf("status: %d, message: %s", err.StatusCode, err.Message)
}

func (c *APIGatewayProxyContext) Bind(v interface{}) error {
	err := json.Unmarshal([]byte(c.Request.Body), v)
	if err != nil {
		return err
	}

	if validatable, ok := v.(Validatable); ok {
		return validatable.Validate()
	}

	return nil
}

func (c *APIGatewayProxyContext) JSON(statusCode int, body interface{}) error {
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

type APIGatewayProxyErrorHandler func(c *APIGatewayProxyContext, err error)

func defaultAPIGatewayProxyErrorHandler(c *APIGatewayProxyContext, err error) {
	var apiGatewayErr APIGatewayProxyError
	switch err := err.(type) {
	case APIGatewayProxyError:
		apiGatewayErr = err
	default:
		apiGatewayErr = APIGatewayProxyError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}
	_ = c.JSON(apiGatewayErr.StatusCode, apiGatewayErr)
}
