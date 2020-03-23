// lmdrouter is a simple-to-use library for writing AWS Lambda functions in Go
// that listen to events of type API Gateway Proxy Request (represented by the
// `events.APIGatewayProxyRequest` type of the github.com/aws-lambda-go/events
// package).
//
// The library allows creating functions that can match requests based on their
// URI, just like an HTTP server that uses the standard
// https://golang.org/pkg/net/http/#ServeMux (or any other community-built routing
// library such as https://github.com/julienschmidt/httprouter or
// https://github.com/go-chi/chi) would. The interface provided by the library
// is very similar to these libraries and should be familiar to anyone who has
// written HTTP applications in Go.
//
// The following features are currently provided:
//
// * Supports all HTTP methods.
//
// * Supports middleware functions at a global and per-resource level.
//
// * Provides ability to automatically "unmarshal" an API Gateway request to an
// arbitrary Go struct, with data coming either from path and query string
// parameters, or from the request body (only JSON requests are currently
// supported). See the documentation for the `UnmarshalRequest` function for
// more information.
//
// * Provides the ability to automatically "marshal" responses of any type to an
// API Gateway response (only JSON responses are currently generated). See the
// MarshalResponse function for more information.
//
package lmdrouter

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type hasMiddleware struct {
	middleware []Middleware
}

// Router is the main type of the library. Lambda routes are registered to it,
// and it's Handler method is used by the lambda to match requests and execute
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
	handler Handler
	hasMiddleware
}

// Middleware is a function that receives a handler function (the next function
// in the chain, possibly another middleware or the actual handler matched for
// a request), and returns a handler function. These functions are quite similar
// to HTTP middlewares in other libraries.
//
// Example middleware that logs all requests:
//
//     func loggerMiddleware(next lmdrouter.Handler) lmdrouter.Handler {
//         return func(ctx context.Context, req events.APIGatewayProxyRequest) (
//             res events.APIGatewayProxyResponse,
//             err error,
//         ) {
//             format := "[%s] [%s %s] [%d]%s"
//             level := "INF"
//             var code int
//             var extra string
//
//             res, err = next(ctx, req)
//             if err != nil {
//                 level = "ERR"
//                 code = http.StatusInternalServerError
//                 extra = " " + err.Error()
//             } else {
//                 code = res.StatusCode
//                 if code >= 400 {
//                     level = "ERR"
//                 }
//             }
//
//             log.Printf(format, level, req.HTTPMethod, req.Path, code, extra)
//
//             return res, err
//         }
//     }
//
type Middleware func(Handler) Handler

// Handler is a request handler function. It receives a context, and the API
// Gateway's proxy request object, and returns a proxy response object and an
// error.
//
// Example:
//
//     func listSomethings(ctx context.Context, req events.APIGatewayProxyRequest) (
//         res events.APIGatewayProxyResponse,
//         err error,
//     ) {
//         // parse input
//         var input listSomethingsInput
//         err = lmdrouter.UnmarshalRequest(req, false, &input)
//         if err != nil {
//             return lmdrouter.HandleError(err)
//         }
//
//         // call some business logic that generates an output struct
//         // ...
//
//         return lmdrouter.MarshalResponse(http.StatusOK, nil, output)
//     }
//
type Handler func(context.Context, events.APIGatewayProxyRequest) (
	events.APIGatewayProxyResponse,
	error,
)

// NewRouter creates a new Router object with a base path and a list of zero or
// more global middleware functions. The base path is necessary if the lambda
// function is not going to be mounted to a domain's root (for example, if the
// function is mounted to "https://my.app/api", then the base path must be
// "/api"). Use an empty string if the function is mounted to the root of the
// domain.
func NewRouter(basePath string, middleware ...Middleware) (l *Router) {
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
func (l *Router) Route(method, path string, handler Handler, middleware ...Middleware) {
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

	r.methods[method] = resource{
		handler: handler,
		hasMiddleware: hasMiddleware{
			middleware: middleware,
		},
	}

	l.routes[path] = r
}

// Handler receives a context and an API Gateway Proxy request, and handles the
// request, matching the appropriate handler and executing it. This is the
// method that must be provided to the lambda's `main` function:
//
//     package main
//
//     import (
//         "github.com/aws/aws-lambda-go/lambda"
//         "github.com/aquasecurity/lmdrouter"
//     )
//
//     var router *lmdrouter.Router
//
//     func init() {
//         router = lmdrouter.NewRouter("/api", loggerMiddleware, authMiddleware)
//         router.Route("GET", "/", listSomethings)
//         router.Route("POST", "/", postSomething, someOtherMiddleware)
//         router.Route("GET", "/:id", getSomething)
//         router.Route("PUT", "/:id", updateSomething)
//         router.Route("DELETE", "/:id", deleteSomething)
//     }
//
//     func main() {
//         lambda.Start(router.Handler)
//     }
//
func (l *Router) Handler(
	ctx context.Context,
	req events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	rsrc, err := l.matchRequest(&req)
	if err != nil {
		return HandleError(err)
	}

	handler := rsrc.handler

	for i := len(rsrc.middleware) - 1; i >= 0; i-- {
		handler = rsrc.middleware[i](handler)
	}
	for i := len(l.middleware) - 1; i >= 0; i-- {
		handler = l.middleware[i](handler)
	}

	return handler(ctx, req)
}

func (l *Router) matchRequest(req *events.APIGatewayProxyRequest) (
	rsrc resource,
	err error,
) {
	// remove trailing slash from request path
	req.Path = strings.TrimSuffix(req.Path, "/")

	// find a route that matches the request
	for _, r := range l.routes {
		// does the path match?
		matches := r.re.FindStringSubmatch(req.Path)
		if matches == nil {
			continue
		}

		// do we have this method?
		var ok bool
		rsrc, ok = r.methods[req.HTTPMethod]
		if !ok {
			return rsrc, HTTPError{
				Code:    http.StatusMethodNotAllowed,
				Message: fmt.Sprintf("%s requests not supported by this resource", req.HTTPMethod),
			}
		}

		// process path parameters
		for i, param := range r.paramNames {
			if len(matches)-1 < len(r.paramNames) {
				return rsrc, HTTPError{
					Code:    http.StatusInternalServerError,
					Message: "Failed matching path parameters",
				}
			}

			if req.PathParameters == nil {
				req.PathParameters = make(map[string]string)
			}

			req.PathParameters[param] = matches[i+1]
		}

		return rsrc, nil
	}

	return rsrc, HTTPError{
		Code:    http.StatusNotFound,
		Message: "No such resource",
	}
}
