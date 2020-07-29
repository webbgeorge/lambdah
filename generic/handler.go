package generic

import (
	"context"
	"encoding/json"

	"github.com/webbgeorge/lambdah"

	"github.com/aws/aws-lambda-go/lambda"
)

type Context struct {
	Context  context.Context
	Event    []byte
	Response interface{}
}

func (c *Context) Bind(v interface{}) error {
	err := json.Unmarshal(c.Event, v)
	if err != nil {
		return err
	}

	if validatable, ok := v.(lambdah.Validatable); ok {
		return validatable.Validate()
	}

	return nil
}

type HandlerFunc func(c *Context) error

func (hf HandlerFunc) Start() {
	lambda.Start(hf.ToLambdaHandler())
}

// Apply middleware to the handler func.
//
// Middleware is called in the order it is given to this function.
func (hf HandlerFunc) Middleware(middleware ...Middleware) HandlerFunc {
	// apply middleware in reverse order
	for i := len(middleware) - 1; i >= 0; i-- {
		hf = middleware[i](hf)
	}
	return hf
}

// Get the AWS Lambda handler of the handler func.
//
// Useful if you need to call AWS lambda.Start(...) directly,
// not required in most cases.
func (hf HandlerFunc) ToLambdaHandler() func(ctx context.Context, event []byte) (interface{}, error) {
	return func(ctx context.Context, event []byte) (interface{}, error) {
		c := &Context{
			Context: ctx,
			Event:   event,
		}
		err := hf(c)
		return c.Response, err
	}
}
