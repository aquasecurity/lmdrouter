// Package lrtr is a simple-to-use library for writing AWS Lambda functions in Go
// that listen to events of type API Gateway Proxy Req (represented by the
// `events.APIGatewayProxyRequest` type of the github.com/aws-lambda-go/events
// package).
//
// The library allows creating functions that can match reqs based on their
// URI, just like an HTTP server that uses the standard
// https://golang.org/pkg/net/http/#ServeMux (or any other community-built routing
// library such as https://github.com/julienschmidt/httprouter or
// https://github.com/go-chi/chi) would. The interface provided by the library
// is very similar to these libraries and should be familiar to anyone who has
// written HTTP applications in Go.
//
// # Use Case
//
// When building large cloud-native applications, there's a certain balance to
// strike when it comes to deployment of APIs. On one side of the scale, each API
// endpoint has its own lambda function. This provides the greatest flexibility,
// but is extremely difficult to maintain. On the other side of the scale, there
// can be one lambda function for the entire API. This provides the least flexibility,
// but is the easiest to maintain. Both are probably not a good idea.
//
// With `lmdrouter`, one can create small lambda functions for different aspects of
// the API. For example, if your application model contains multiple domains (e.g.
// articles, authors, topics, etc...), you can create one lambda function for each
// domain, and deploy these independently (e.g. everything below "/api/articles" is
// one lambda function, everything below "/api/authors" is another function). This
// is also useful for applications where different teams are in charge of different
// parts of the API.
//
// # Features
//
// * Supports all HTTP methods.
//
// * Supports middleware functions at a global and per-resource level.
//
// * Supports path parameters with a simple ":<name>" format (e.g. "/posts/:id").
//
// * Provides ability to automatically "unmarshal" an API Gateway req to an
// arbitrary Go struct, with data coming either from path and query string
// parameters, or from the req body (only JSON reqs are currently
// supported). See the documentation for the `UnmarshalReq` function for
// more information.
//
// * Provides the ability to automatically "marshal" responses of any type to an
// API Gateway response (only JSON responses are currently generated). See the
// CustomRes function for more information.
//
//   - Implements net/http.Handler for local development and general usage outside
//     an AWS Lambda environment.
package lrtr

import (
	"context"
	"fmt"
	"github.com/seantcanavan/lambda_jwt_router/lcom"
	"github.com/seantcanavan/lambda_jwt_router/lmw"
	"github.com/seantcanavan/lambda_jwt_router/lres"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type hasMiddleware struct {
	middleware []lcom.Middleware
}

// Router is the main type of the library. Lambda routes are registered to it,
// and it's Handler method is used by the lambda to match reqs and execute
// the appropriate handler.
type Router struct {
	basePath string
	routes   map[string]route
	hasMiddleware
}

type route struct {
	re         *regexp.Regexp
	paramNames []string
	methods    map[string]resource
}

type resource struct {
	handler lcom.Handler
	hasMiddleware
}

// NewRouter creates a new Router object with a base path and a list of zero or
// more global middleware functions. The base path is necessary if the lambda
// function is not going to be mounted to a domain's root (for example, if the
// function is mounted to "https://my.app/api", then the base path must be
// "/api"). Use an empty string if the function is mounted to the root of the
// domain.
func NewRouter(basePath string, middleware ...lcom.Middleware) (l *Router) {
	return &Router{
		basePath: basePath,
		routes:   make(map[string]route),
		hasMiddleware: hasMiddleware{
			middleware: middleware,
		},
	}
}

// Route registers a new route, with the provided HTTP method name and path,
// and zero or more local middleware functions.
func (l *Router) Route(method, path string, handler lcom.Handler, middleware ...lcom.Middleware) {
	// check if this route already exists
	r, ok := l.routes[path]
	if !ok {
		r = route{
			methods: make(map[string]resource),
		}

		// create a regular expression from the path
		var err error
		re := fmt.Sprintf("^%s", l.basePath)
		for _, part := range strings.Split(path, "/") {
			if part == "" {
				continue
			}

			// is this a parameter?
			if strings.HasPrefix(part, ":") {
				r.paramNames = append(r.paramNames, strings.TrimPrefix(part, ":"))
				re = fmt.Sprintf("%s/([^/]+)", re)
			} else {
				re = fmt.Sprintf("%s/%s", re, part)
			}
		}
		re = fmt.Sprintf("%s$", re)

		r.re, err = regexp.Compile(re)
		if err != nil {
			panic(fmt.Sprintf("Generated invalid regex for route %s", path))
		}
	}

	// unless CORS is overridden - we place an options handler at the
	// current route. If the new method/route is OPTIONS then the
	// code after this will override it with a new OPTIONS handler. If
	// this isn't an OPTIONS call, this will add support for CORS
	// for that specific route. No middleware can be applied here
	// for simplicity reasons as any middleware that performs
	// authentication or authorization on the main route will also
	// apply here and prevent the CORS request from succeeding.
	if os.Getenv(lcom.NoCORS) != "true" {
		r.methods[http.MethodOptions] = resource{
			handler: lmw.AllowOptionsMW(),
		}
	}

	r.methods[method] = resource{
		handler: handler,
		hasMiddleware: hasMiddleware{
			middleware: middleware,
		},
	}

	l.routes[path] = r
}

// Handler receives a context and an API Gateway Proxy req, and handles the
// req, matching the appropriate handler and executing it. This is the
// method that must be provided to the lambda's `main` function:
//
//	package main
//
//	import (
//	    "github.com/aws/aws-lambda-go/lambda"
//	    "github.com/seantcanavan/lmdrouter"
//	)
//
//	var router *lmdrouter.Router
//
//	func init() {
//	    router = lmdrouter.NewRouter("/api", loggerMiddleware, authMiddleware)
//	    router.Route(http.MethodGet, "/", listSomethings)
//	    router.Route(http.MethodPost, "/", postSomething, someOtherMiddleware)
//	    router.Route(http.MethodGet, "/:id", getSomething)
//	    router.Route("PUT", "/:id", updateSomething)
//	    router.Route(http.MethodDelete, "/:id", deleteSomething)
//	}
//
//	func main() {
//	    lambda.Start(router.Handler)
//	}
func (l *Router) Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	matchedResource, err := l.matchReq(&req)
	if err != nil {
		return lres.ErrorRes(err)
	}

	handler := matchedResource.handler

	for i := len(matchedResource.middleware) - 1; i >= 0; i-- {
		handler = matchedResource.middleware[i](handler)
	}
	for i := len(l.middleware) - 1; i >= 0; i-- {
		handler = l.middleware[i](handler)
	}

	return handler(ctx, req)
}

func (l *Router) matchReq(req *events.APIGatewayProxyRequest) (
	matchedResource resource,
	err error,
) {
	// remove trailing slash from req path
	req.Path = strings.TrimSuffix(req.Path, "/")

	negErr := lres.HTTPError{
		Status:  http.StatusNotFound,
		Message: "No such resource",
	}

	// find a route that matches the req
	for _, r := range l.routes {
		// does the path match?
		matches := r.re.FindStringSubmatch(req.Path)
		if matches == nil {
			continue
		}

		// do we have this method?
		var ok bool
		matchedResource, ok = r.methods[req.HTTPMethod]
		if !ok {
			// we matched a route, but it didn't support this method. Mark negErr
			// with a 405 error, but continue, we might match another route
			negErr = lres.HTTPError{
				Status:  http.StatusMethodNotAllowed,
				Message: fmt.Sprintf("%s reqs not supported by this resource", req.HTTPMethod),
			}
			continue
		}

		// process path parameters
		for i, param := range r.paramNames {
			if len(matches)-1 < len(r.paramNames) {
				return matchedResource, lres.HTTPError{
					Status:  http.StatusInternalServerError,
					Message: "Failed matching path parameters",
				}
			}

			if req.PathParameters == nil {
				req.PathParameters = make(map[string]string)
			}

			req.PathParameters[param], _ = url.QueryUnescape(matches[i+1])
		}

		return matchedResource, nil
	}

	return matchedResource, negErr
}
