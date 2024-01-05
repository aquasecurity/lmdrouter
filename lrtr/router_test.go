package lrtr

import (
	"context"
	"errors"
	"fmt"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/seantcanavan/lambda_jwt_router/lcom"
	"github.com/seantcanavan/lambda_jwt_router/lreq"
	"github.com/seantcanavan/lambda_jwt_router/lres"
	"github.com/stretchr/testify/require"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

var testLog []string

func TestRouter(t *testing.T) {
	lmd := NewRouter("/api", logger)
	lmd.Route(http.MethodGet, "/", listSomethings)
	lmd.Route(http.MethodPost, "/", postSomething, auth)
	lmd.Route(http.MethodGet, "/:id", getSomething)
	lmd.Route(http.MethodGet, "/:id/stuff", listStuff)
	lmd.Route(http.MethodGet, "/:id/stuff/:fake", listStuff)

	t.Run("Routes created correctly", func(t *testing.T) {
		t.Run("/", func(t *testing.T) {
			route, ok := lmd.routes["/"]
			require.True(t, ok, "Route must be created")
			if ok {
				require.Equal(t, `^/api$`, route.re.String(), "Regex must be correct")
				require.NotEqual(t, nil, route.methods[http.MethodGet], "GET method must exist")
				require.NotEqual(t, nil, route.methods[http.MethodPost], "POST method must exist")
				require.NotEqual(t, nil, route.methods[http.MethodOptions], "OPTIONS method must exist") // auto generated for CORS support
			}
		})
		t.Run("/:id", func(t *testing.T) {
			route, ok := lmd.routes["/:id"]
			require.True(t, ok, "Route must be created")
			if ok {
				require.Equal(t, `^/api/([^/]+)$`, route.re.String(), "Regex must be correct")
				require.NotEqual(t, nil, route.methods[http.MethodGet], "GET method must exist")
				require.NotEqual(t, nil, route.methods[http.MethodOptions], "OPTIONS method must exist") // auto generated for CORS support
			}
		})
		t.Run("/:id/stuff/:fake", func(t *testing.T) {
			route, ok := lmd.routes["/:id/stuff/:fake"]
			require.True(t, ok, "Route must be created")
			if ok {
				require.Equal(
					t,
					`^/api/([^/]+)/stuff/([^/]+)$`,
					route.re.String(),
					"Regex must be correct",
				)
				require.EqualValues(
					t,
					[]string{"id", "fake"},
					route.paramNames,
					"Param names must be correct",
				)
			}
		})
	})

	t.Run("Reqs matched correctly", func(t *testing.T) {
		t.Run("POST /api", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Path:       "/api",
			}
			_, err := lmd.matchReq(&req)
			require.Equal(t, nil, err, "ErrorRes must be nil")
		})

		t.Run("POST /api/", func(t *testing.T) {
			// make sure trailing slashes are removed
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Path:       "/api/",
			}
			_, err := lmd.matchReq(&req)
			require.Equal(t, nil, err, "ErrorRes must be nil")
		})

		t.Run("DELETE /api", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodDelete,
				Path:       "/api",
			}
			_, err := lmd.matchReq(&req)
			require.NotEqual(t, nil, err, "ErrorRes must not be nil")
			var httpErr lres.HTTPError
			ok := errors.As(err, &httpErr)
			require.True(t, ok, "ErrorRes must be an HTTP error")
			require.Equal(t, http.StatusMethodNotAllowed, httpErr.Status, "ErrorRes code must be 405")
		})

		t.Run("GET /api/fake-id", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
				Path:       "/api/fake-id",
			}
			_, err := lmd.matchReq(&req)
			require.Equal(t, nil, err, "ErrorRes must be nil")
			require.Equal(t, "fake-id", req.PathParameters["id"], "ID must be correct")
		})

		t.Run("GET /api/fake-id/bla", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
				Path:       "/api/fake-id/bla",
			}
			_, err := lmd.matchReq(&req)
			require.NotEqual(t, nil, err, "ErrorRes must not be nil")
			var httpErr lres.HTTPError
			ok := errors.As(err, &httpErr)
			require.True(t, ok, "ErrorRes must be an HTTP error")
			require.Equal(t, http.StatusNotFound, httpErr.Status, "ErrorRes code must be 404")
		})

		t.Run("GET /api/fake-id/stuff/faked-fake", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
				Path:       "/api/fake-id/stuff/faked-fake",
			}
			_, err := lmd.matchReq(&req)
			require.Equal(t, nil, err, "ErrorRes must be nil")
			require.Equal(t, "fake-id", req.PathParameters["id"], "'id' must be correct")
			require.Equal(t, "faked-fake", req.PathParameters["fake"], "'fake' must be correct")
		})
	})

	t.Run("Reqs execute correctly", func(t *testing.T) {
		t.Run("POST /api without auth", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Path:       "/api",
			}
			res, err := lmd.Handler(context.Background(), req)
			require.Equal(t, nil, err, "ErrorRes must not be nil")
			require.Equal(t, http.StatusUnauthorized, res.StatusCode, "Status code must be 401")
			require.True(t, len(testLog) > 0, "Log must have items")
			require.Equal(
				t,
				"[ERR] [POST /api] [401]",
				testLog[len(testLog)-1],
				"Last long line must be correct",
			)
		})

		t.Run("POST /api with auth", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Path:       "/api",
				Headers: map[string]string{
					"Authorization": "Bearer fake-token",
				},
			}
			res, err := lmd.Handler(context.Background(), req)
			require.Equal(t, nil, err, "ErrorRes must not be nil")
			require.Equal(t, http.StatusBadRequest, res.StatusCode, "Status code must be 400")
		})

		t.Run("GET /api", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodGet,
				Path:       "/api",
			}
			res, err := lmd.Handler(context.Background(), req)
			require.Equal(t, nil, err, "ErrorRes must not be nil")
			require.Equal(t, http.StatusOK, res.StatusCode, "Status code must be 200")
			require.True(t, len(testLog) > 0, "Log must have items")
			require.Equal(
				t,
				"[INF] [GET /api] [200]",
				testLog[len(testLog)-1],
				"Last long line must be correct",
			)
		})
	})

	t.Run("Overlapping routes", func(t *testing.T) {
		router := NewRouter("")
		router.Route(
			http.MethodGet,
			"/foo/:id",
			func(_ context.Context, _ events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
				res.Body = "/foo/:id"
				return res, nil
			},
		)
		router.Route(
			http.MethodPost,
			"/foo/bar",
			func(_ context.Context, _ events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
				res.Body = "/foo/bar"
				return res, nil
			},
		)

		// call POST /foo/bar in a loop. We do this because the router iterates
		// over a map to match routes, which is non-deterministic, meaning
		// sometimes we may match the route and sometimes not
		for i := 1; i <= 10; i++ {
			res, _ := router.Handler(context.Background(), events.APIGatewayProxyRequest{
				HTTPMethod: http.MethodPost,
				Path:       "/foo/bar",
			})
			require.Equal(t, "/foo/bar", res.Body, "req must match /foo/bar route")
		}

		res, _ := router.Handler(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodDelete,
			Path:       "/foo/bar",
		})
		require.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, "Status code must be 405")

		res, _ = router.Handler(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodGet,
			Path:       "/foo/bar2",
		})
		require.Equal(t, "/foo/:id", res.Body, "Body must match")
	})
}

func listSomethings(_ context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	// parse input
	var input util.MockListReq
	err = lreq.UnmarshalReq(req, false, &input)
	if err != nil {
		return lres.ErrorRes(err)
	}

	now := time.Now()
	then := now.Add(-time.Hour * 32)

	output := []util.MockItem{
		{ID: "one", Name: "First Item", Date: now},
		{ID: "two", Name: "2nd Item", Date: then},
		{ID: "three", Name: "Third Item", Date: then},
	}

	return lres.CustomRes(http.StatusOK, nil, output)
}

func postSomething(_ context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	var input util.MockPostReq
	err = lreq.UnmarshalReq(req, true, &input)
	if err != nil {
		return lres.ErrorRes(err)
	}

	output := map[string]string{
		"id":  "bla",
		"url": "https://service.com/api/bla",
	}

	return lres.CustomRes(http.StatusAccepted, map[string]string{
		"Location": output["url"],
	}, output)
}

func getSomething(_ context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	// parse input
	var input util.MockGetReq
	err = lreq.UnmarshalReq(req, false, &input)
	if err != nil {
		return lres.ErrorRes(err)
	}

	output := util.MockItem{
		ID:   input.ID,
		Name: "Fake Name",
		Date: time.Now(),
	}

	return lres.CustomRes(http.StatusOK, nil, output)
}

func listStuff(_ context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	// parse input
	var input util.MockListReq
	err = lreq.UnmarshalReq(req, false, &input)
	if err != nil {
		return lres.ErrorRes(err)
	}

	output := make([]util.MockItem, len(input.Terms))
	for i, term := range input.Terms {
		output[i] = util.MockItem{
			ID:   input.ID,
			Name: fmt.Sprintf("%s in %s", term, input.Language),
		}
	}

	return lres.CustomRes(http.StatusOK, nil, output)
}

func logger(next lcom.Handler) lcom.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		// [LEVEL] [METHOD PATH] [CODE] EXTRA
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

		testLog = append(testLog, fmt.Sprintf(
			format,
			level,
			req.HTTPMethod,
			req.Path,
			code,
			extra,
		))

		return res, err
	}
}

func auth(next lcom.Handler) lcom.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		auth := req.Headers["Authorization"]
		if auth != "" && strings.HasPrefix(auth, "Bearer ") {
			token := strings.TrimPrefix(auth, "Bearer ")
			if token == "fake-token" {
				return next(ctx, req)
			}
		}

		return lres.CustomRes(
			http.StatusUnauthorized,
			map[string]string{"WWW-Authenticate": "Bearer"},
			lres.HTTPError{Status: http.StatusUnauthorized, Message: "Unauthorized"},
		)
	}
}
