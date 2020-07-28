package main

import (
	"io"
	"os"

	lambdah "github.com/webbgeorge/lambdah/cloudwatch_events"
	"github.com/webbgeorge/lambdah/log"
)

func main() {
	newHandler(os.Stdout).Start()
}

// example: just log the name from the event detail
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

		return nil
	}).Middleware(
		lambdah.CorrelationIDMiddleware(),
		lambdah.LoggerMiddleware(logWriter, map[string]string{
			"appName":      "lambdahExamples",
			"functionName": "cloudWatchEventsMiddlewareExample",
		}),
	)
}

type eventData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}
