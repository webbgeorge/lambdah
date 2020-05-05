package lambdah

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestAPIGatewayProxyHandler_Success(t *testing.T) {
	h := func(c *APIGatewayProxyContext) error {
		var data requestData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		assert.Equal(t, "hello", data.Message)

		return c.JSON(http.StatusOK, responseData{Status: "all good"})
	}

	awsHandler := APIGatewayProxyHandler(h)
	res, err := awsHandler(
		context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, `{"status":"all good"}`, res.Body)
}

func TestAPIGatewayProxyHandler_ErrorIsHandled(t *testing.T) {
	h := func(c *APIGatewayProxyContext) error {
		return errors.New("an error happened")
	}

	awsHandler := APIGatewayProxyHandler(h)
	res, err := awsHandler(
		context.Background(),
		events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	)

	assert.Nil(t, err)
	assert.Equal(t, 500, res.StatusCode)
	assert.Equal(t, `{"message":"Internal server error"}`, res.Body)
}

func TestAPIGatewayProxyContext_Bind_Success(t *testing.T) {
	c := &APIGatewayProxyContext{
		Request: events.APIGatewayProxyRequest{
			Body: `{"message": "hello"}`,
		},
	}

	var data requestData
	err := c.Bind(&data)

	assert.Nil(t, err)
	assert.Equal(t, "hello", data.Message)
}

func TestAPIGatewayProxyContext_Bind_InvalidJSON(t *testing.T) {
	c := &APIGatewayProxyContext{
		Request: events.APIGatewayProxyRequest{
			Body: `{"messag`,
		},
	}

	var data requestData
	err := c.Bind(&data)

	assert.Error(t, err)
}

func TestAPIGatewayProxyContext_Bind_WithValidationError(t *testing.T) {
	c := &APIGatewayProxyContext{
		Request: events.APIGatewayProxyRequest{
			Body: `{}`,
		},
	}

	var data requestData
	err := c.Bind(&data)

	assert.NotNil(t, err)
	assert.Equal(t, "invalid message", err.Error())
}

func TestAPIGatewayProxyContext_JSON_WithBody(t *testing.T) {
	c := &APIGatewayProxyContext{}

	err := c.JSON(http.StatusOK, responseData{Status: "all good"})

	assert.Nil(t, err)
	assert.Equal(t, 200, c.Response.StatusCode)
	assert.Equal(t, `{"status":"all good"}`, c.Response.Body)
}

func TestAPIGatewayProxyContext_JSON_WithoutBody(t *testing.T) {
	c := &APIGatewayProxyContext{}

	err := c.JSON(http.StatusNoContent, nil)

	assert.Nil(t, err)
	assert.Equal(t, 204, c.Response.StatusCode)
	assert.Equal(t, ``, c.Response.Body)
}

func TestAPIGatewayProxyContext_HandleError_UnhandledError(t *testing.T) {
	c := &APIGatewayProxyContext{}

	c.handleError(errors.New("some error"))

	assert.Equal(t, 500, c.Response.StatusCode)
	assert.Equal(t, `{"message":"Internal server error"}`, c.Response.Body)
}

func TestAPIGatewayProxyContext_HandleError_APIGatewayProxyError(t *testing.T) {
	c := &APIGatewayProxyContext{}

	c.handleError(APIGatewayProxyError{
		StatusCode: 400,
		Message:    "Bad request",
	})

	assert.Equal(t, 400, c.Response.StatusCode)
	assert.Equal(t, `{"message":"Bad request"}`, c.Response.Body)
}

type requestData struct {
	Message string `json:"message"`
}

func (d *requestData) Validate() error {
	if d.Message == "" {
		return errors.New("invalid message")
	}
	return nil
}

type responseData struct {
	Status string `json:"status"`
}
