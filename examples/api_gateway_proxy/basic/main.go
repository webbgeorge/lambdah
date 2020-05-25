package main

import (
	"fmt"
	"net/http"

	"github.com/webbgeorge/lambdah/api_gateway_proxy"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	h := api_gateway_proxy.Handler(api_gateway_proxy.HandlerConfig{}, newHandler())
	lambda.Start(h)
}

func newHandler() api_gateway_proxy.HandlerFunc {
	return func(c *api_gateway_proxy.Context) error {
		var data requestData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		if data.Name == "Dave" {
			return api_gateway_proxy.Error{
				StatusCode: http.StatusNotAcceptable,
				Message:    "Dave is not welcome here!",
			}
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
		return api_gateway_proxy.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "Greeting not allowed",
		}
	}
	if d.Name == "" {
		return api_gateway_proxy.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "Name is required",
		}
	}
	return nil
}

type responseData struct {
	Message string `json:"message"`
}
