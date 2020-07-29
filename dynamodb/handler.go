package dynamodb

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type Context struct {
	Context     context.Context
	EventRecord events.DynamoDBEventRecord
}

func (c *Context) BindKeys(out interface{}) error {
	return unmarshalStreamImage(c.EventRecord.Change.Keys, out)
}

func (c *Context) BindNewImage(out interface{}) error {
	return unmarshalStreamImage(c.EventRecord.Change.NewImage, out)
}

func (c *Context) BindOldImage(out interface{}) error {
	return unmarshalStreamImage(c.EventRecord.Change.OldImage, out)
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
func (hf HandlerFunc) ToLambdaHandler() func(ctx context.Context, event events.DynamoDBEvent) error {
	return func(ctx context.Context, event events.DynamoDBEvent) error {
		for _, eventRecord := range event.Records {
			c := &Context{
				Context:     ctx,
				EventRecord: eventRecord,
			}
			err := hf(c)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func unmarshalStreamImage(attributes map[string]events.DynamoDBAttributeValue, out interface{}) error {
	sdkAttributeMap := make(map[string]*dynamodb.AttributeValue)

	for k, attribute := range attributes {
		var sdkAttribute dynamodb.AttributeValue

		bytes, err := attribute.MarshalJSON()
		if err != nil {
			return err
		}

		err = json.Unmarshal(bytes, &sdkAttribute)
		if err != nil {
			return err
		}

		sdkAttributeMap[k] = &sdkAttribute
	}

	return dynamodbattribute.UnmarshalMap(sdkAttributeMap, out)
}
