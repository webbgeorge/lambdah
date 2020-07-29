package main

import (
	"errors"
	"io"
	"os"

	lambdah "github.com/webbgeorge/lambdah/generic"
)

func main() {
	newHandler(os.Stdout).Start()
}

// example: just log the name from the event detail and respond "ok"
func newHandler(logger io.Writer) lambdah.HandlerFunc {
	return func(c *lambdah.Context) error {
		var data eventData
		err := c.Bind(&data)
		if err != nil {
			return err
		}

		_, _ = logger.Write([]byte(data.Name))

		c.Response = "ok"

		return nil
	}
}

type eventData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}

func (d *eventData) Validate() error {
	if d.Greeting != "Hi" && d.Greeting != "Hello" {
		return errors.New("greeting not allowed")
	}
	if d.Name == "" {
		return errors.New("name is required")
	}
	return nil
}
