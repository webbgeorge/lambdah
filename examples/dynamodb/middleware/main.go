package main

import (
	"io"
	"os"

	lambdah "github.com/webbgeorge/lambdah/dynamodb"
	"github.com/webbgeorge/lambdah/log"
)

func main() {
	newHandler(os.Stdout).Start()
}

// log event names to context logger
func newHandler(logWriter io.Writer) lambdah.HandlerFunc {
	return lambdah.HandlerFunc(func(c *lambdah.Context) error {
		log.LoggerFromContext(c.Context).
			Info().
			Msgf("Event name: '%s'", c.EventRecord.EventName)
		return nil
	}).Middleware(
		lambdah.CorrelationIDMiddleware(),
		lambdah.LoggerMiddleware(logWriter, map[string]string{
			"appName":      "lambdahExamples",
			"functionName": "dynamodbMiddlewareExample",
		}),
	)
}
