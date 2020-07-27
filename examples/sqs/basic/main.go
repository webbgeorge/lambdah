package main

import (
	"errors"
	"io"
	"os"

	lambdah "github.com/webbgeorge/lambdah/sqs"
)

func main() {
	newHandler(os.Stdout).Start()
}

// example: just log the name from the message
func newHandler(logger io.Writer) lambdah.HandlerFunc {
	return func(c *lambdah.Context) error {
		var data messageData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		_, _ = logger.Write([]byte(data.Name))

		return nil
	}
}

type messageData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}

func (d *messageData) Validate() error {
	if d.Greeting != "Hi" && d.Greeting != "Hello" {
		return errors.New("greeting not allowed")
	}
	if d.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
