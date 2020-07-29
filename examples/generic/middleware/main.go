package main

import (
	"io"
	"os"

	lambdah "github.com/webbgeorge/lambdah/generic"
	"github.com/webbgeorge/lambdah/log"
)

func main() {
	newHandler(os.Stdout).Start()
}

// example: just log the name from the event detail and respond "ok"
func newHandler(logWriter io.Writer) lambdah.HandlerFunc {
	return lambdah.HandlerFunc(func(c *lambdah.Context) error {
		var data eventData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		log.LoggerFromContext(c.Context).
			Info().
			Msgf("name received: %s", data.Name)

		c.Response = "ok"

		return nil
	}).Middleware(
		lambdah.CorrelationIDMiddleware(),
		lambdah.LoggerMiddleware(logWriter, map[string]string{
			"appName":      "lambdahExamples",
			"functionName": "genericMiddlewareExample",
		}),
	)
}

type eventData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}
