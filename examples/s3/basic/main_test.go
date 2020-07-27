package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/stretchr/testify/assert"
)

func TestNewHandler_Success_SkipOtherEvents(t *testing.T) {
	mock := &s3Mock{}

	h := newHandler(mock, ioutil.Discard).ToLambdaHandler()

	err := h(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{{EventName: "s3:test:event"}}},
	)

	assert.Nil(t, err)
	assert.False(t, mock.GetObjectInvoked)
}

func TestNewHandler_Success_CreatedEvent(t *testing.T) {
	mockS3Client := &s3Mock{
		GetObjectMock: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			assert.Equal(t, aws.String("my-bucket"), input.Bucket)
			assert.Equal(t, aws.String("my-key"), input.Key)
			return &s3.GetObjectOutput{
				Body: ioutil.NopCloser(bytes.NewBufferString("line one\nline two")),
			}, nil
		},
	}
	mockLogger := &bytes.Buffer{}

	h := newHandler(mockS3Client, mockLogger).ToLambdaHandler()

	err := h(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{{
			EventName: "ObjectCreated:Put",
			S3: events.S3Entity{
				Bucket: events.S3Bucket{
					Name: "my-bucket",
				},
				Object: events.S3Object{
					Key: "my-key",
				},
			},
		}}},
	)

	assert.Nil(t, err)
	assert.True(t, mockS3Client.GetObjectInvoked)
	assert.Equal(t, "line one", mockLogger.String())
}

func TestNewHandler_Failure_S3Error(t *testing.T) {
	mockS3Client := &s3Mock{
		GetObjectMock: func(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
			return nil, assert.AnError
		},
	}

	h := newHandler(mockS3Client, ioutil.Discard).ToLambdaHandler()

	err := h(
		context.Background(),
		events.S3Event{Records: []events.S3EventRecord{{
			EventName: "ObjectCreated:Put",
			S3: events.S3Entity{
				Bucket: events.S3Bucket{
					Name: "my-bucket",
				},
				Object: events.S3Object{
					Key: "my-key",
				},
			},
		}}},
	)

	assert.Equal(t, assert.AnError, err)
	assert.True(t, mockS3Client.GetObjectInvoked)
}

type s3Mock struct {
	s3.S3
	GetObjectInvoked bool
	GetObjectMock    func(*s3.GetObjectInput) (*s3.GetObjectOutput, error)
}

var _ s3iface.S3API = &s3Mock{}

func (s *s3Mock) GetObject(in *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	s.GetObjectInvoked = true
	return s.GetObjectMock(in)
}
