package middleware

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/seantcanavan/lmdrouter"
	"github.com/seantcanavan/lmdrouter/response"
	"log"
	"net/http"
)

// AllowOptions is a helper middleware function that will immediately return
// a successful request if the method is OPTIONS. This makes sure that
// HTTP OPTIONS calls for CORS functionality are supported.
func AllowOptions(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		if req.HTTPMethod == "OPTIONS" { // immediately return success for options calls for CORS reqs
			return response.Empty()
		}

		return next(ctx, req)
	}
}

// LogRequest is a standard middleware function that will log every incoming
// events.APIGatewayProxyRequest request and the pertinent information in it.
func LogRequest(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		format := "[%s] [%s %s] [%d]%s"
		level := "INF"
		var code int
		var extra string

		res, err = next(ctx, req)
		if err != nil {
			level = "ERR"
			code = http.StatusInternalServerError
			extra = " " + err.Error()
		} else {
			code = res.StatusCode
			if code >= 400 {
				level = "ERR"
			}
		}

		log.Printf(format, level, req.HTTPMethod, req.Path, code, extra)

		return res, err
	}
}
