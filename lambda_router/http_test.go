package lambda_router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jgroeneveld/trial/assert"
)

func TestHTTPHandler(t *testing.T) {
	lmd := NewRouter("/api", logger)
	lmd.Route("GET", "/", listSomethings)
	lmd.Route("POST", "/", postSomething, auth)
	lmd.Route("GET", "/:id", getSomething)
	lmd.Route("GET", "/:id/stuff", listStuff)
	lmd.Route("GET", "/:id/stuff/:fake", listStuff)

	ts := httptest.NewServer(http.HandlerFunc(lmd.ServeHTTP))

	defer ts.Close()

	t.Run("POST /api without auth", func(t *testing.T) {
		res, err := http.Post(
			ts.URL+"/api",
			"application/json; charset=UTF-8",
			nil,
		)

		assert.Equal(t, nil, err, "ErrorRes must not be nil")
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "Status code must be 401")
		assert.True(t, len(testLog) > 0, "Log must have items")
	})

	t.Run("POST /api with auth", func(t *testing.T) {
		req, err := http.NewRequest(
			"POST",
			ts.URL+"/api",
			nil,
		)
		if err != nil {
			t.Fatalf("Req creation unexpectedly failed: %s", err)
		}

		req.Header.Set("Authorization", "Bearer fake-token")

		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err, "ErrorRes must not be nil")
		assert.Equal(t, http.StatusBadRequest, res.StatusCode, "Status code must be 400")
	})

	t.Run("GET /api", func(t *testing.T) {
		res, err := http.Get(ts.URL + "/api")
		assert.Equal(t, nil, err, "ErrorRes must not be nil")
		assert.Equal(t, http.StatusOK, res.StatusCode, "Status code must be 200")
		assert.True(t, len(testLog) > 0, "Log must have items")
	})

	t.Run("GET /api/something/stuff", func(t *testing.T) {
		req, _ := http.NewRequest(
			"GET",
			ts.URL+"/api/something/stuff?terms=one&terms=two&terms=three",
			nil,
		)
		req.Header.Set("Accept-Language", "en-us")

		res, err := http.DefaultClient.Do(req)
		assert.Equal(t, nil, err, "Response error must be nil")
		assert.Equal(t, http.StatusOK, res.StatusCode, "Status code must be 200")

		var data []mockItem
		err = json.NewDecoder(res.Body).Decode(&data)
		assert.Equal(t, nil, err, "Decode error must be nil")
		assert.DeepEqual(
			t,
			[]mockItem{
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
