package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/webbgeorge/lambdah"
	"net/http"
)

func main() {
	h := lambdah.APIGatewayProxyHandler(lambdah.APIGatewayProxyHandlerConfig{}, newHandler())
	lambda.Start(h)
}

func newHandler() lambdah.APIGatewayProxyHandlerFunc {
	return func(c *lambdah.APIGatewayProxyContext) error {
		var data requestData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		message := fmt.Sprintf("%s %s", data.Greeting, data.Name)

		return c.JSON(http.StatusOK, responseData{Message: message})
	}
}

type requestData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}

func (d *requestData) Validate() error {
	if d.Greeting != "Hi" && d.Greeting != "Hello" {
		return lambdah.APIGatewayProxyError{
			StatusCode: http.StatusBadRequest,
			Message:    "Greeting not allowed",
		}
	}
	if d.Name == "" {
		return lambdah.APIGatewayProxyError{
			StatusCode: http.StatusBadRequest,
			Message:    "Name is required",
		}
	}
	return nil
}

type responseData struct {
	Message string `json:"message"`
}
