package lambdah

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// TODO: logging?
// TODO: allowing users to define error type/structure

type APIGatewayHandlerFunc func(c *APIGatewayContext) error

func APIGatewayHandler(h APIGatewayHandlerFunc) func(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context,
		request events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		c := &APIGatewayContext{
			Context: ctx,
			Request: request,
		}

		err := h(c)
		if err != nil {
			c.handleError(err)
			return c.Response, nil
		}

		return c.Response, nil
	}
}

type APIGatewayContext struct {
	Context  context.Context
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

type Validatable interface {
	Validate() error
}

type APIGatewayError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (err APIGatewayError) Error() string {
	return fmt.Sprintf("status: %d, message: %s", err.StatusCode, err.Message)
}

func (c *APIGatewayContext) Bind(v interface{}) error {
	err := json.Unmarshal([]byte(c.Request.Body), v)
	if err != nil {
		return err
	}

	if validatable, ok := v.(Validatable); ok {
		return validatable.Validate()
	}

	return nil
}

func (c *APIGatewayContext) JSON(statusCode int, body interface{}) error {
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

func (c *APIGatewayContext) handleError(err error) {
	var apiGatewayErr APIGatewayError
	switch err := err.(type) {
	case APIGatewayError:
		apiGatewayErr = err
	default:
		apiGatewayErr = APIGatewayError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}
	_ = c.JSON(apiGatewayErr.StatusCode, apiGatewayErr)
}

// Example code:
func exampleHandler() APIGatewayHandlerFunc {
	return func(c *APIGatewayContext) error {
		var requestData requestBody
		err := c.Bind(&requestData)
		if err != nil {
			return err
		}

		err := doThingOne(requestData)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responseBody{Status: "error one"})
		}

		err := doThingTwo()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, responseBody{Status: "error two"})
		}

		return c.JSON(http.StatusOK, responseBody{Status: "ok"})
	}
}

type requestBody struct {
	FieldOne string `json:"field_one"`
	FieldTwo int    `json:"field_two"`
}

func (b *requestBody) Validate() error {
	if b.FieldOne == "" {
		return APIGatewayError{
			StatusCode: http.StatusBadRequest,
			Message:    "field one is empty",
		}
	}
	if b.FieldTwo > 0 && b.FieldTwo < 10 {
		return APIGatewayError{
			StatusCode: http.StatusBadRequest,
			Message:    "field two must be between 1 and 10",
		}
	}
	return nil
}

type responseBody struct {
	Status string `json:"status"`
}
