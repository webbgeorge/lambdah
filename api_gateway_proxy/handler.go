package api_gateway_proxy

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/gorilla/reverse"
)

// TODO: logging?

type Context struct {
	Context  context.Context
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

type Validatable interface {
	Validate() error
}

func (c *Context) Bind(v interface{}) error {
	err := json.Unmarshal([]byte(c.Request.Body), v)
	if err != nil {
		return err
	}

	if validatable, ok := v.(Validatable); ok {
		return validatable.Validate()
	}

	return nil
}

func (c *Context) JSON(statusCode int, body interface{}) error {
	if body != nil {
		if c.Response.Headers == nil {
			c.Response.Headers = make(map[string]string)
		}
		c.Response.Headers["Content-Type"] = "application/json"

		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		c.Response.Body = string(b)
	}
	c.Response.StatusCode = statusCode
	return nil
}

// Lambdah API Gateway Proxy handler function
type HandlerFunc func(c *Context) error

// Starts the lambda function.
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
func (hf HandlerFunc) ToLambdaHandler() func(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context,
		request events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		c := &Context{
			Context: ctx,
			Request: request,
		}

		err := hf(c)
		if err != nil {
			// catch any unhandled errors and return default error
			// if error handler middleware is on, no errors will return here
			c.Response.StatusCode = http.StatusInternalServerError
			c.Response.Body = "Internal server error"

			if c.Response.Headers == nil {
				c.Response.Headers = make(map[string]string)
			}
			c.Response.Headers["Content-Type"] = "text/html"

			return c.Response, nil
		}

		return c.Response, nil
	}
}

// ToHTTPHandler turns the Lambdah handler function into a go http.Handler.
// This is useful for using go http testing tools with API gateway proxy handler.
//
// resourcePathPattern is a path pattern used to simulate AWS API Gateway path matching
// functionality for use within your tests. E.g. `/books/{bookID}` would be available from
// the context as c.Request.PathParameters["bookID"]
//
// stageVariables are to simulate any stage variables you have set up on your api
// gateway in aws
func (hf HandlerFunc) ToHttpHandler(
	resourcePathPattern string,
	stageVariables map[string]string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		proxyResponse, err := hf.ToLambdaHandler()(r.Context(), events.APIGatewayProxyRequest{
			Resource:                        resourcePathPattern,
			Path:                            r.URL.Path,
			HTTPMethod:                      r.Method,
			Headers:                         singleValue(r.Header),
			MultiValueHeaders:               r.Header,
			QueryStringParameters:           singleValue(r.URL.Query()),
			MultiValueQueryStringParameters: r.URL.Query(),
			PathParameters:                  parsePathParams(resourcePathPattern, r.URL.Path),
			StageVariables:                  stageVariables,
			Body:                            string(body),
		})

		if err != nil {
			// write a generic error, the same as API GW would if an error was returned by handler
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`error`))
			return
		}

		writeResponse(w, proxyResponse)
	})
}

func singleValue(multiValueMap map[string][]string) map[string]string {
	singleValueMap := make(map[string]string)
	for k, mv := range multiValueMap {
		if len(mv) > 0 {
			singleValueMap[k] = mv[0]
		}
	}
	return singleValueMap
}

func parsePathParams(pathPattern string, path string) map[string]string {
	re, err := reverse.NewGorillaPath(pathPattern, false)
	if err != nil {
		return map[string]string{}
	}

	params := make(map[string]string)
	if matches := re.MatchString(path); matches {
		for name, values := range re.Values(path) {
			if len(values) > 0 {
				params[name] = values[0]
			}
		}
	}

	return params
}

func writeResponse(w http.ResponseWriter, proxyResponse events.APIGatewayProxyResponse) {
	for k, v := range proxyResponse.Headers {
		w.Header().Add(k, v)
	}

	for k, vs := range proxyResponse.MultiValueHeaders {
		for _, v := range vs {
			w.Header().Add(k, v)
		}
	}

	w.WriteHeader(proxyResponse.StatusCode)
	_, _ = w.Write([]byte(proxyResponse.Body))
}
