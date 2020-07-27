package s3

import (
	"context"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

func TestS3Handler_Success_OneEvent(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		assert.Equal(t, "s3:test:event", c.EventRecord.EventName)
		assert.Equal(t, "testBucket", c.EventRecord.S3.Bucket.Name)
		assert.Equal(t, "test.txt", c.EventRecord.S3.Object.Key)
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{{
			EventName: "s3:test:event",
			S3: events.S3Entity{
				Bucket: events.S3Bucket{
					Name: "testBucket",
				},
				Object: events.S3Object{
					Key: "test.txt",
				},
			},
		}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, 1, callCount)
}

func TestS3Handler_Success_MultipleEvents(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		switch callCount {
		case 1:
			assert.Equal(t, "s3:test:event:1", c.EventRecord.EventName)
			assert.Equal(t, "testBucket1", c.EventRecord.S3.Bucket.Name)
			assert.Equal(t, "test1.txt", c.EventRecord.S3.Object.Key)
		case 2:
			assert.Equal(t, "s3:test:event:2", c.EventRecord.EventName)
			assert.Equal(t, "testBucket2", c.EventRecord.S3.Bucket.Name)
			assert.Equal(t, "test2.txt", c.EventRecord.S3.Object.Key)
		}
		return nil
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{
			{
				EventName: "s3:test:event:1",
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: "testBucket1",
					},
					Object: events.S3Object{
						Key: "test1.txt",
					},
				},
			},
			{
				EventName: "s3:test:event:2",
				S3: events.S3Entity{
					Bucket: events.S3Bucket{
						Name: "testBucket2",
					},
					Object: events.S3Object{
						Key: "test2.txt",
					},
				},
			},
		}},
	)

	assert.Nil(t, err)
	assert.Equal(t, 2, callCount)
}

func TestS3Handler_Error(t *testing.T) {
	callCount := 0
	h := func(c *Context) error {
		callCount++
		return assert.AnError
	}

	awsHandler := HandlerFunc(h).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{{EventName: "s3:test:event"}}},
	)

	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, 1, callCount)
}

func TestS3Handler_Middleware(t *testing.T) {
	callOrder := make([]string, 0)

	h := func(c *Context) error {
		callOrder = append(callOrder, "handler")
		return nil
	}

	mw1 := func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			callOrder = append(callOrder, "mw1 in")
			err := h(c)
			callOrder = append(callOrder, "mw1 out")
			return err
		}
	}

	mw2 := func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			callOrder = append(callOrder, "mw2 in")
			err := h(c)
			callOrder = append(callOrder, "mw2 out")
			return err
		}
	}

	awsHandler := HandlerFunc(h).Middleware(mw1, mw2).ToLambdaHandler()
	err := awsHandler(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{{EventName: "s3:test:event"}}},
	)

	assert.Nil(t, err)
	assert.Equal(t, []string{"mw1 in", "mw2 in", "handler", "mw2 out", "mw1 out"}, callOrder)
}
