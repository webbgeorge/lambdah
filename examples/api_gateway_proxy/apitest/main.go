package main

import (
	"errors"
	"net/http"

	lambdah "github.com/webbgeorge/lambdah/api_gateway_proxy"
)

func main() {
	newHandler().Start()
}

func newHandler() lambdah.HandlerFunc {
	return lambdah.HandlerFunc(func(c *lambdah.Context) error {
		switch c.Request.PathParameters["name"] {
		case "giraffe":
			return c.JSON(http.StatusOK, responseData{
				Animal: "Giraffe",
				Trait:  "tall",
			})
		case "mouse":
			return c.JSON(http.StatusOK, responseData{
				Animal: "Mouse",
				Trait:  "small",
			})
		case "cat":
			return errors.New("cat causes 500")
		default:
			return lambdah.Error{
				StatusCode: 400,
				Message:    "Animal not found",
			}
		}
	}).Middleware(lambdah.ErrorHandlerMiddleware())
}

type responseData struct {
	Animal string `json:"animal"`
	Trait  string `json:"trait"`
}
