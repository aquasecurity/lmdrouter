package lrtr

import (
	"encoding/json"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTPHandler(t *testing.T) {
	lmd := NewRouter("/api", logger)
	lmd.Route(http.MethodGet, "/", listSomethings)
	lmd.Route(http.MethodPost, "/", postSomething, auth)
	lmd.Route(http.MethodGet, "/:id", getSomething)
	lmd.Route(http.MethodGet, "/:id/stuff", listStuff)
	lmd.Route(http.MethodGet, "/:id/stuff/:fake", listStuff)

	ts := httptest.NewServer(http.HandlerFunc(lmd.ServeHTTP))

	defer ts.Close()

	t.Run("POST /api without auth", func(t *testing.T) {
		res, err := http.Post(
			ts.URL+"/api",
			"application/json; charset=UTF-8",
			nil,
		)

		require.Equal(t, nil, err, "ErrorRes must not be nil")
		require.Equal(t, http.StatusUnauthorized, res.StatusCode, "Status code must be 401")
		require.True(t, len(testLog) > 0, "Log must have items")
	})

	t.Run("POST /api with auth", func(t *testing.T) {
		req, err := http.NewRequest(
			http.MethodPost,
			ts.URL+"/api",
			nil,
		)
		if err != nil {
			t.Fatalf("Req creation unexpectedly failed: %s", err)
		}

		req.Header.Set("Authorization", "Bearer fake-token")

		res, err := http.DefaultClient.Do(req)
		require.Equal(t, nil, err, "ErrorRes must not be nil")
		require.Equal(t, http.StatusBadRequest, res.StatusCode, "Status code must be 400")
	})

	t.Run("GET /api", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/api")
		require.Equal(t, nil, err, "ErrorRes must not be nil")
		require.Equal(t, http.StatusOK, res.StatusCode, "Status code must be 200")
		require.True(t, len(testLog) > 0, "Log must have items")
	})

	t.Run("GET /api/something/stuff", func(t *testing.T) {
		req, _ := http.NewRequest(
			http.MethodGet,
			ts.URL+"/api/something/stuff?terms=one&terms=two&terms=three",
			nil,
		)
		req.Header.Set("Accept-Language", "en-us")

		res, err := http.DefaultClient.Do(req)
		require.Equal(t, nil, err, "Response error must be nil")
		require.Equal(t, http.StatusOK, res.StatusCode, "Status code must be 200")

		var data []util.MockItem
		err = json.NewDecoder(res.Body).Decode(&data)
		require.Equal(t, nil, err, "Decode error must be nil")
		require.EqualValues(
			t,
			[]util.MockItem{
				{
					ID:   "something",
					Name: "one in en-us",
					Date: time.Time{},
				},
				{
					ID:   "something",
					Name: "two in en-us",
					Date: time.Time{},
				},
				{
					ID:   "something",
					Name: "three in en-us",
					Date: time.Time{},
				},
			},
			data,
			"Response body must match",
		)
	})
}
