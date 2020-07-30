# lambdah

[![Godoc](https://godoc.org/github.com/webbgeorge/lambdah?status.svg)](https://pkg.go.dev/github.com/webbgeorge/lambdah)
![CI](https://github.com/webbgeorge/lambdah/workflows/CI/badge.svg)
[![Coverage Status](https://coveralls.io/repos/github/webbgeorge/lambdah/badge.svg?branch=master)](https://coveralls.io/github/webbgeorge/lambdah?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/webbgeorge/lambdah)](https://goreportcard.com/report/github.com/webbgeorge/lambdah)

**lambdah** provides a useful abstraction layer over AWS Lambda functions written in Go.

Features:

* A simple interface for writing lambda functions of all types
* Extensible middleware framework
* Data binding for JSON payloads
* Highly customisable
* Optional logging and tracing tools via middleware

## Installation

**lambdah** is available as a [Go module](https://github.com/golang/go/wiki/Modules) - import
any package under `github.com/webbgeorge/lambdah` in your project to get started.

## Getting started

**lambdah** has handlers for many types of event. Here are two example handlers for API
Gateway Proxy Requests and SQS Messages.

### API Gateway Proxy request handler

```go
package main

import (
	"net/http"

	lambdah "github.com/webbgeorge/lambdah/api_gateway_proxy"
)

func main() {
	lambdah.HandlerFunc(handler).Start()
}

// handle http request via API Gateway
func handler(c *lambdah.Context) error {
	return c.String(http.StatusOK, "Hello world!")
}
```

### SQS message handler

```go
package main

import (
	"fmt"

	lambdah "github.com/webbgeorge/lambdah/sqs"
)

func main() {
	lambdah.HandlerFunc(handler).Start()
}

// handle SQS message
func handler(c *lambdah.Context) error {
	// Unmarshal JSON message data into a struct
	var data messageData
	err := c.Bind(&data)
	if err != nil {
		return err
	}

	// just log the name to CloudWatch Logs (via stdout) for the example
	fmt.Println(data.Name)

	return nil
}

type messageData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}
```

## More examples

Handler type      | Example
----------------- | -----------------
api_gateway_proxy | [basic](examples/api_gateway_proxy/basic)
api_gateway_proxy | [custom error handler](examples/api_gateway_proxy/custom_error_handler)
api_gateway_proxy | [apitest](examples/api_gateway_proxy/apitest)
cloudwatch_events | [basic](examples/cloudwatch_events/basic)
cloudwatch_events | [middleware](examples/cloudwatch_events/middleware)
dynamodb          | [basic](examples/dynamodb/basic)
dynamodb          | [middleware](examples/dynamodb/middleware)
generic           | [basic](examples/generic/basic)
generic           | [middleware](examples/generic/middleware)
s3                | [basic](examples/s3/basic)
s3                | [middleware](examples/s3/middleware)
sns               | [basic](examples/sns/basic)
sns               | [middleware](examples/sns/middleware)
sqs               | [basic](examples/sqs/basic)
sqs               | [middleware](examples/sqs/middleware)

## Concepts

### Middleware

**lambdah** provides a middleware system to allow extending the functionality of its handlers.
Using middleware on any handler is completely optional.

Each handler includes some default middleware that you can choose to use or not. This
includes a `CorrelationIDMiddleware` and a `LoggerMiddleware` for all handlers. Some
handlers have additional default middleware, such as the
`api_gateway_proxy.ErrorHandlerMiddleware` middleware.

In addition to default middleware, custom middleware can be created using a simple API,
described below.

#### Using middleware

Here is the same simple API Gateway Proxy handler from above, with the addition
of two middleware.

```go
package main

import (
	"net/http"
	"os"

	lambdah "github.com/webbgeorge/lambdah/api_gateway_proxy"
)

func main() {
	lambdah.
		HandlerFunc(handler).
		Middleware(
			lambdah.CorrelationIDMiddleware(),
			lambdah.LoggerMiddleware(os.Stdout, nil),
		).
		Start()
}

// handle http request via API Gateway
func handler(c *lambdah.Context) error {
	return c.String(http.StatusOK, "Hello world!")
}
```

#### Custom middleware

Here is a simple example of a custom middleware for the API Gateway Proxy handler.

```go
package main

import (
	"fmt"

	lambdah "github.com/webbgeorge/lambdah/api_gateway_proxy"
)

// function which returns our custom middleware
func MyCustomMiddleware() lambdah.Middleware {
	return func(h lambdah.HandlerFunc) lambdah.HandlerFunc {
		return func(c *lambdah.Context) error {
			// just log to CloudWatch Logs (via stdout) for the example
			fmt.Println("middleware triggered")
			return h(c)
		}
	}
}
```

#### Default middleware

Handler           | Default middleware
----------------- | ------------------
api_gateway_proxy | ErrorHandlerMiddleware, CorrelationIDMiddleware, LoggerMiddleware
cloudwatch_events | CorrelationIDMiddleware, LoggerMiddleware
dynamodb          | CorrelationIDMiddleware, LoggerMiddleware
generic           | CorrelationIDMiddleware, LoggerMiddleware
s3                | CorrelationIDMiddleware, LoggerMiddleware
sns               | CorrelationIDMiddleware, LoggerMiddleware
sqs               | CorrelationIDMiddleware, LoggerMiddleware

### Binding data

Many of the handler types allow you to bind payload data into a Go data structure.

This includes:

* API Gateway Proxy request body (from JSON)
* CloudWatch event detail (from JSON)
* DynamoDB event images (from DynamoDB attribute map)
* Generic event payload (from JSON)
* SNS event data (from JSON)
* SQS message data (from JSON)

For example, here is an example SNS event handler which binds JSON data:

```go
package main

import (
	"fmt"

	lambdah "github.com/webbgeorge/lambdah/sns"
)

func main() {
	lambdah.HandlerFunc(handler).Start()
}

// handle SNS event
func handler(c *lambdah.Context) error {
	// Unmarshal JSON message data into a struct
	var data messageData
	err := c.Bind(&data)
	if err != nil {
		return err
	}

	// just log the name to CloudWatch Logs (via stdout) for the example
	fmt.Println(data.Name)

	return nil
}

type messageData struct {
	Greeting string `json:"greeting"`
	Name     string `json:"name"`
}
```

### Logging

**lambdah** provides some built-in logging support using middleware. Logging can be 
an opinionated matter, so using this is completely optional.

To enable a logger on any type of handler, attach its `LoggerMiddleware`. This
attaches a [zerolog](https://github.com/rs/zerolog) logger to the context of
each event. Two arguments are required to use the logger middleware:
 
* an `io.Writer` to write log messages to
* a `map[string]string` of fields to add to the log data (set to nil if not wanted)

```go
lambdah.
	HandlerFunc(handler).
	Middleware(
		lambdah.LoggerMiddleware(os.Stdout, map[string]string{"logFieldOne": "valueOne"}),
	).
	Start()
```

The logger middleware will, by default, log the status of each processed event with
some basic details about the event. The logger can also be accessed within your
handlers and custom middleware.

```go
func handler(c *lambdah.Context) error {
	logger := log.LoggerFromContext(c.Context)
	logger.Info().Msg("hello logs")
	return c.String(http.StatusOK, "Hello world!")
}
```

### Correlation IDs

**lambdah** provides an optional `CorrelationIDMiddleware` for all of its handlers.
This middleware retrieves or creates a Correlation ID, which is often used for
tracing requests through distributed systems in logs. This correlation ID is
always a v4 UUID e.g. `9187bcbb-052b-45e5-bffa-99f323a8aa71`.

When used in conjunction with the `LoggerMiddleware`, each log message includes
the Correlation ID. Note that the `CorrelationIDMiddleware` must be included
before the `LoggerMiddleware`.

The Correlation ID can also be accessed from within handlers and custom middleware:

```go
func handler(c *lambdah.Context) error {
	cid := log.CorrelationIDFromContext(c.Context)
	// log correlation id for example
	fmt.Println(cid)
	return c.String(http.StatusOK, "Hello world!")
}
```

#### Source of Correlation ID

The handler sources or creates the Correlation ID in different ways depending 
on its type.

Handler           | Correlation ID Source
----------------- | ------------------
api_gateway_proxy | `Correlation-Id` request header if present, otherwise is created by lambdah
cloudwatch_events | CloudWatch Event ID
dynamodb          | created by lambdah
generic           | created by lambdah
s3                | created by lambdah
sns               | `correlation_id` SNS message attribute if present, otherwise is created by lambdah
sqs               | `correlation_id` SQS message attribute if present, otherwise is created by lambdah

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md)
