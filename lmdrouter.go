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

type Middleware func(Handler) Handler

type Handler func(context.Context, events.APIGatewayProxyRequest) (
	events.APIGatewayProxyResponse,
	error,
)

func NewRouter(basePath string, middleware ...Middleware) (l *Router) {
	return &Router{
		basePath: basePath,
		routes:   make(map[string]route),
		hasMiddleware: hasMiddleware{
			middleware: middleware,
		},
	}
}

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
