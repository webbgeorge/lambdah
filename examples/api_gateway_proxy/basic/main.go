package main

import (
	"fmt"
	"net/http"

	lambdah "github.com/webbgeorge/lambdah/api_gateway_proxy"
)

func main() {
	newHandler().Start()
}

func newHandler() lambdah.HandlerFunc {
	return lambdah.HandlerFunc(func(c *lambdah.Context) error {
		var data requestData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		if data.Name == "Dave" {
			return lambdah.Error{
				StatusCode: http.StatusNotAcceptable,
				Message:    "Dave is not welcome here!",
			}
		}

		message := fmt.Sprintf("%s %s", data.Greeting, data.Name)

		return c.JSON(http.StatusOK, responseData{Message: message})
	}).Middleware(lambdah.ErrorHandlerMiddleware())
}

type requestData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}

func (d *requestData) Validate() error {
	if d.Greeting != "Hi" && d.Greeting != "Hello" {
		return lambdah.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "Greeting not allowed",
		}
	}
	if d.Name == "" {
		return lambdah.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "Name is required",
		}
	}
	return nil
}

type responseData struct {
	Message string `json:"message"`
}
