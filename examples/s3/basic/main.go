package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	lambdah "github.com/webbgeorge/lambdah/s3"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

func main() {
	awsConf, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	scClient := s3.New(awsConf)

	newHandler(scClient, os.Stdout).Start()
}

// on ObjectCreated events, download the file and log first line
func newHandler(
	s3Client s3iface.S3API,
	logger io.Writer,
) lambdah.HandlerFunc {
	return func(c *lambdah.Context) error {
		if !strings.HasPrefix(c.EventRecord.EventName, "ObjectCreated:") {
			// Skipping all other events
			return nil
		}

		output, err := s3Client.GetObject(&s3.GetObjectInput{
			Bucket: aws.String(c.EventRecord.S3.Bucket.Name),
			Key:    aws.String(c.EventRecord.S3.Object.Key),
		})
		if err != nil {
			return err
		}

		scanner := bufio.NewScanner(output.Body)

		// scan just one line and print it
		scanner.Scan()
		_, _ = logger.Write([]byte(scanner.Text()))
		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	}
}
