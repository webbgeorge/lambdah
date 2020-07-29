package main

import (
	"fmt"
	"io"
	"os"

	lambdah "github.com/webbgeorge/lambdah/dynamodb"
)

func main() {
	newHandler(os.Stdout).Start()
}

// example: on dynamodb stream events, log data
func newHandler(logger io.Writer) lambdah.HandlerFunc {
	return func(c *lambdah.Context) error {
		var d data
		err := c.BindNewImage(&d)
		if err != nil {
			return err
		}

		_, _ = logger.Write([]byte(fmt.Sprintf(
			"dynamodb change of type '%s' for name '%s'",
			c.EventRecord.EventName,
			d.Name,
		)))

		return nil
	}
}

type data struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}
