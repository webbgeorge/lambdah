package main

import (
	"io"
	"os"

	"github.com/webbgeorge/lambdah/log"
	lambdah "github.com/webbgeorge/lambdah/sns"
)

func main() {
	newHandler(os.Stdout).Start()
}

// example: just log the name from the message
func newHandler(logWriter io.Writer) lambdah.HandlerFunc {
	return lambdah.HandlerFunc(func(c *lambdah.Context) error {
		var data messageData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		log.LoggerFromContext(c.Context).
			Info().
			Msgf("name received: %s", data.Name)

		return nil
	}).Middleware(
		lambdah.CorrelationIDMiddleware(),
		lambdah.LoggerMiddleware(logWriter, map[string]string{
			"appName":      "lambdahExamples",
			"functionName": "snsMiddlewareExample",
		}),
	)
}

type messageData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}
