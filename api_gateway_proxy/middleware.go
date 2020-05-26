package api_gateway_proxy

import "github.com/google/uuid"

type Middleware func(h HandlerFunc) HandlerFunc

// Middleware to get or attach a correlation ID to the request, useful for tracing
// requests through distributed systems.
//
// First looks for a Correlation ID provided in the request header `Correlation-Id`.
// If not present a new correlation ID will be created and put into the request header.
func CorrelationIDMiddleware() Middleware {
	return func(h HandlerFunc) HandlerFunc {
		return func(c *Context) error {
			cid := c.Request.Headers["Correlation-Id"]
			if cid == "" {
				cid = uuid.New().String()

				if c.Request.Headers == nil {
					c.Request.Headers = make(map[string]string)
				}
				c.Request.Headers["Correlation-Id"] = cid

				if c.Request.MultiValueHeaders == nil {
					c.Request.MultiValueHeaders = make(map[string][]string)
				}
				c.Request.MultiValueHeaders["Correlation-Id"] = []string{cid}
			}
			return h(c)
		}
	}
}
