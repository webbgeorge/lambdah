package s3

import (
	"io"

	"github.com/webbgeorge/lambdah/log"
)

type Middleware func(h HandlerFunc) HandlerFunc

// Middleware to attach a correlation ID to the event, useful for tracing
// requests through distributed systems if it is passed along in further requests.
// You can also access the correlation ID directly in your handlers and
// middlewares by calling log.CorrelationIDFromContext(c.Context).
//
// If used with the LoggerMiddleware, Correlation IDs are logged in each message.
// To ensure logs have correlation ID field, this middleware should be called first.
func CorrelationIDMiddleware() Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			// TODO: review correlation ID input
			// TODO: implement correlation IDs from metadata when AWS Lambda event includes them
			c.Context = log.WithCorrelationID(c.Context, log.NewCorrelationID())
			return h(c)
		}
	}
}

// Middleware to configure a logger in the context.Context.
//
// The middleware logs on each event handled, and on errors. In addition to
// the default log messages, you can access the logger in your handlers/middleware
// by calling log.LoggerFromContext(c.Context). The logger we use is
// github.com/rs/zerolog
//
// w io.Writer              is the log output, for example os.Stdout
// field map[string]string  is a list of key value fields to include in each log message
func LoggerMiddleware(w io.Writer, fields map[string]string) Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			fields["handler_type"] = "s3"
			fields["correlation_id"] = log.CorrelationIDFromContext(c.Context)
			fields["event_name"] = c.EventRecord.EventName
			fields["bucket_name"] = c.EventRecord.S3.Bucket.Name
			fields["object_key"] = c.EventRecord.S3.Object.Key

			logger := log.NewLogger(w, fields)
			c.Context = log.WithLogger(c.Context, logger)
			logger.Info().Msgf("Processing S3 event '%s'", c.EventRecord.EventName)
			err := h(c)
			if err != nil {
				logger.Error().
					Msgf("Error processing S3 event: %s", err.Error())
			}
			return err
		}
	}
}
