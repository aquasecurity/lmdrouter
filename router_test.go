package lmdrouter

import (
	"context"
	"errors"
	"fmt"
	"github.com/seantcanavan/lmdrouter/response"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jgroeneveld/trial/assert"
)

var testLog []string

func TestRouter(t *testing.T) {
	lmd := NewRouter("/api", logger)
	lmd.Route("GET", "/", listSomethings)
	lmd.Route("POST", "/", postSomething, auth)
	lmd.Route("GET", "/:id", getSomething)
	lmd.Route("GET", "/:id/stuff", listStuff)
	lmd.Route("GET", "/:id/stuff/:fake", listStuff)

	t.Run("Routes created correctly", func(t *testing.T) {
		t.Run("/", func(t *testing.T) {
			route, ok := lmd.routes["/"]
			assert.True(t, ok, "Route must be created")
			if ok {
				assert.Equal(t, `^/api$`, route.re.String(), "Regex must be correct")
				assert.NotEqual(t, nil, route.methods["GET"], "GET method must exist")
				assert.NotEqual(t, nil, route.methods["POST"], "POST method must exist")
			}
		})

		t.Run("/:id", func(t *testing.T) {
			route, ok := lmd.routes["/:id"]
			assert.True(t, ok, "Route must be created")
			if ok {
				assert.Equal(t, `^/api/([^/]+)$`, route.re.String(), "Regex must be correct")
				assert.NotEqual(t, nil, route.methods["GET"], "GET method must exist")
			}
		})

		t.Run("/:id/stuff/:fake", func(t *testing.T) {
			route, ok := lmd.routes["/:id/stuff/:fake"]
			assert.True(t, ok, "Route must be created")
			if ok {
				assert.Equal(
					t,
					`^/api/([^/]+)/stuff/([^/]+)$`,
					route.re.String(),
					"Regex must be correct",
				)
				assert.DeepEqual(
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
				HTTPMethod: "POST",
				Path:       "/api",
			}
			_, err := lmd.matchReq(&req)
			assert.Equal(t, nil, err, "Error must be nil")
		})

		t.Run("POST /api/", func(t *testing.T) {
			// make sure trailing slashes are removed
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				Path:       "/api/",
			}
			_, err := lmd.matchReq(&req)
			assert.Equal(t, nil, err, "Error must be nil")
		})

		t.Run("DELETE /api", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "DELETE",
				Path:       "/api",
			}
			_, err := lmd.matchReq(&req)
			assert.NotEqual(t, nil, err, "Error must not be nil")
			var httpErr HTTPError
			ok := errors.As(err, &httpErr)
			assert.True(t, ok, "Error must be an HTTP error")
			assert.Equal(t, http.StatusMethodNotAllowed, httpErr.Status, "Error code must be 405")
		})

		t.Run("GET /api/fake-id", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/api/fake-id",
			}
			_, err := lmd.matchReq(&req)
			assert.Equal(t, nil, err, "Error must be nil")
			assert.Equal(t, "fake-id", req.PathParameters["id"], "ID must be correct")
		})

		t.Run("GET /api/fake-id/bla", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/api/fake-id/bla",
			}
			_, err := lmd.matchReq(&req)
			assert.NotEqual(t, nil, err, "Error must not be nil")
			var httpErr HTTPError
			ok := errors.As(err, &httpErr)
			assert.True(t, ok, "Error must be an HTTP error")
			assert.Equal(t, http.StatusNotFound, httpErr.Status, "Error code must be 404")
		})

		t.Run("GET /api/fake-id/stuff/fakey-fake", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/api/fake-id/stuff/fakey-fake",
			}
			_, err := lmd.matchReq(&req)
			assert.Equal(t, nil, err, "Error must be nil")
			assert.Equal(t, "fake-id", req.PathParameters["id"], "'id' must be correct")
			assert.Equal(t, "fakey-fake", req.PathParameters["fake"], "'fake' must be correct")
		})
	})

	t.Run("Reqs execute correctly", func(t *testing.T) {
		t.Run("POST /api without auth", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				Path:       "/api",
			}
			res, err := lmd.Handler(context.Background(), req)
			assert.Equal(t, nil, err, "Error must not be nil")
			assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Status code must be 401")
			assert.True(t, len(testLog) > 0, "Log must have items")
			assert.Equal(
				t,
				"[ERR] [POST /api] [401]",
				testLog[len(testLog)-1],
				"Last long line must be correct",
			)
		})

		t.Run("POST /api with auth", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "POST",
				Path:       "/api",
				Headers: map[string]string{
					"Authorization": "Bearer fake-token",
				},
			}
			res, err := lmd.Handler(context.Background(), req)
			assert.Equal(t, nil, err, "Error must not be nil")
			assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Status code must be 400")
		})

		t.Run("GET /api", func(t *testing.T) {
			req := events.APIGatewayProxyRequest{
				HTTPMethod: "GET",
				Path:       "/api",
			}
			res, err := lmd.Handler(context.Background(), req)
			assert.Equal(t, nil, err, "Error must not be nil")
			assert.Equal(t, http.StatusOK, res.StatusCode, "Status code must be 200")
			assert.True(t, len(testLog) > 0, "Log must have items")
			assert.Equal(
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
			"GET",
			"/foo/:id",
			func(_ context.Context, _ events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse, err error) {
				res.Body = "/foo/:id"
				return res, nil
			},
		)
		router.Route(
			"POST",
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
				HTTPMethod: "POST",
				Path:       "/foo/bar",
			})
			assert.Equal(t, "/foo/bar", res.Body, "req must match /foo/bar route")
		}

		res, _ := router.Handler(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod: "DELETE",
			Path:       "/foo/bar",
		})
		assert.Equal(t, http.StatusMethodNotAllowed, res.StatusCode, "Status code must be 405")

		res, _ = router.Handler(context.Background(), events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Path:       "/foo/bar2",
		})
		assert.Equal(t, "/foo/:id", res.Body, "Body must match")
	})
}

func listSomethings(ctx context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	// parse input
	var input mockListReq
	err = UnmarshalRequest(req, false, &input)
	if err != nil {
		return response.Error(err)
	}

	now := time.Now()
	then := now.Add(-time.Hour * 32)

	output := []mockItem{
		{ID: "one", Name: "First Item", Date: now},
		{ID: "two", Name: "2nd Item", Date: then},
		{ID: "three", Name: "Third Item", Date: then},
	}

	return response.Custom(http.StatusOK, nil, output)
}

func postSomething(ctx context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	var input mockPostReq
	err = UnmarshalRequest(req, true, &input)
	if err != nil {
		return response.Error(err)
	}

	output := map[string]string{
		"id":  "bla",
		"url": "https://service.com/api/bla",
	}

	return response.Custom(http.StatusAccepted, map[string]string{
		"Location": output["url"],
	}, output)
}

func getSomething(ctx context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	// parse input
	var input mockGetReq
	err = UnmarshalRequest(req, false, &input)
	if err != nil {
		return response.Error(err)
	}

	output := mockItem{
		ID:   input.ID,
		Name: "Fake Name",
		Date: time.Now(),
	}

	return response.Custom(http.StatusOK, nil, output)
}

func listStuff(ctx context.Context, req events.APIGatewayProxyRequest) (
	res events.APIGatewayProxyResponse,
	err error,
) {
	// parse input
	var input mockListReq
	err = UnmarshalRequest(req, false, &input)
	if err != nil {
		return response.Error(err)
	}

	output := make([]mockItem, len(input.Terms))
	for i, term := range input.Terms {
		output[i] = mockItem{
			ID:   input.ID,
			Name: fmt.Sprintf("%s in %s", term, input.Language),
		}
	}

	return response.Custom(http.StatusOK, nil, output)
}

func logger(next Handler) Handler {
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

func auth(next Handler) Handler {
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

		return response.Custom(
			http.StatusUnauthorized,
			map[string]string{"WWW-Authenticate": "Bearer"},
			HTTPError{http.StatusUnauthorized, "Unauthorized"},
		)
	}
}
