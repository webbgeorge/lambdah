package lambdah

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/gorilla/reverse"
)

// TODO: logging?

type APIGatewayProxyHandlerFunc func(c *APIGatewayProxyContext) error

type APIGatewayProxyHandlerConfig struct {
	ErrorHandler APIGatewayProxyErrorHandler
}

func APIGatewayProxyHandler(conf APIGatewayProxyHandlerConfig, h APIGatewayProxyHandlerFunc) func(
	ctx context.Context,
	request events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	return func(
		ctx context.Context,
		request events.APIGatewayProxyRequest,
	) (events.APIGatewayProxyResponse, error) {
		c := &APIGatewayProxyContext{
			Context: ctx,
			Request: request,
		}

		err := h(c)
		if err != nil {
			if conf.ErrorHandler == nil {
				defaultAPIGatewayProxyErrorHandler(c, err)
			} else {
				conf.ErrorHandler(c, err)
			}
			return c.Response, nil
		}

		return c.Response, nil
	}
}

type APIGatewayProxyContext struct {
	Context  context.Context
	Request  events.APIGatewayProxyRequest
	Response events.APIGatewayProxyResponse
}

type Validatable interface {
	Validate() error
}

type APIGatewayProxyError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (err APIGatewayProxyError) Error() string {
	return fmt.Sprintf("status: %d, message: %s", err.StatusCode, err.Message)
}

func (c *APIGatewayProxyContext) Bind(v interface{}) error {
	err := json.Unmarshal([]byte(c.Request.Body), v)
	if err != nil {
		return err
	}

	if validatable, ok := v.(Validatable); ok {
		return validatable.Validate()
	}

	return nil
}

func (c *APIGatewayProxyContext) JSON(statusCode int, body interface{}) error {
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		c.Response.Body = string(b)
	}
	c.Response.StatusCode = statusCode
	return nil
}

type APIGatewayProxyErrorHandler func(c *APIGatewayProxyContext, err error)

func defaultAPIGatewayProxyErrorHandler(c *APIGatewayProxyContext, err error) {
	var apiGatewayErr APIGatewayProxyError
	switch err := err.(type) {
	case APIGatewayProxyError:
		apiGatewayErr = err
	default:
		apiGatewayErr = APIGatewayProxyError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Internal server error",
		}
	}
	_ = c.JSON(apiGatewayErr.StatusCode, apiGatewayErr)
}

// AWS API Gateway Proxy Lambda SDK handler function type
type AWSAPIGatewayProxyHandlerFunc func(ctx context.Context, r events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// Turns an AWS API Gateway Proxy Lambda SDK handler function into a go http.Handler
// this is useful for using go http testing tools with API gateway proxy handler
func HttpHandlerFromAWSAPIGatewayProxyHandler(
	awsHandler AWSAPIGatewayProxyHandlerFunc,
	resourcePathPattern string,
	stageVariables map[string]string,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		proxyResponse, err := awsHandler(r.Context(), events.APIGatewayProxyRequest{
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
