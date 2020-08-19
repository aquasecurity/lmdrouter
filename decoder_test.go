package lmdrouter

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jgroeneveld/trial/assert"
)

func Test_UnmarshalRequest(t *testing.T) {
	t.Run("valid path&query input", func(t *testing.T) {
		var input mockListRequest
		err := UnmarshalRequest(
			events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"id": "fake-scan-id",
				},
				QueryStringParameters: map[string]string{
					"page":      "2",
					"page_size": "30",
					"const":     "two",
				},
				MultiValueQueryStringParameters: map[string][]string{
					"terms":   []string{"one", "two"},
					"numbers": []string{"1.2", "3.5", "666.666"},
				},
				Headers: map[string]string{
					"Accept-Language": "en-us",
				},
				MultiValueHeaders: map[string][]string{
					"Accept-Encoding": []string{"gzip", "deflate"},
				},
			},
			false,
			&input,
		)
		assert.Equal(t, nil, err, "Error must be nil")
		assert.Equal(t, "fake-scan-id", input.ID, "ID must be parsed from path")
		assert.Equal(t, int64(2), input.Page, "Page must be parsed from query")
		assert.Equal(t, int64(30), input.PageSize, "PageSize must be parsed from query")
		assert.Equal(t, "en-us", input.Language, "Language must be parsed from headers")
		assert.Equal(t, mockConstTwo, input.Const, "Const must be parsed from query")
		assert.DeepEqual(t, []string{"one", "two"}, input.Terms, "Terms must be parsed from multiple query params")
		assert.DeepEqual(t, []float64{1.2, 3.5, 666.666}, input.Numbers, "Numbers must be parsed from multiple query params")
		assert.DeepEqual(t, []string{"gzip", "deflate"}, input.Encoding, "Encoding must be parsed from multiple header params")
	})

	t.Run("invalid path&query input", func(t *testing.T) {
		var input mockListRequest
		err := UnmarshalRequest(
			events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"id": "fake-scan-id",
				},
				QueryStringParameters: map[string]string{
					"page": "abcd",
				},
			},
			false,
			&input,
		)
		assert.NotEqual(t, nil, err, "Error must not be nil")
		var httpErr HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Error must be an HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code, "Error code must be 400")
	})

	fakeDate := time.Date(2020, 3, 23, 11, 33, 0, 0, time.UTC)

	t.Run("valid body input, not base64", func(t *testing.T) {
		var input mockPostRequest
		err := UnmarshalRequest(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: false,
				PathParameters: map[string]string{
					"id": "bla",
				},
				Body: `{"name":"Fake Post","date":"2020-03-23T11:33:00Z"}`,
			},
			true,
			&input,
		)

		assert.Equal(t, nil, err, "Error must be nil")
		assert.Equal(t, "bla", input.ID, "ID must be parsed from path parameters")
		assert.Equal(t, "Fake Post", input.Name, "Name must be parsed from body")
		assert.Equal(t, fakeDate, input.Date, "Date must be parsed from body")
	})

	t.Run("invalid body input, not base64", func(t *testing.T) {
		var input mockPostRequest
		err := UnmarshalRequest(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: false,
				Body:            `this is not JSON`,
			},
			true,
			&input,
		)

		assert.NotEqual(t, nil, err, "Error must not be nil")
	})

	t.Run("valid body input, base64", func(t *testing.T) {
		var input mockPostRequest
		err := UnmarshalRequest(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: true,
				Body:            "eyJuYW1lIjoiRmFrZSBQb3N0IiwiZGF0ZSI6IjIwMjAtMDMtMjNUMTE6MzM6MDBaIn0=",
			},
			true,
			&input,
		)

		assert.Equal(t, nil, err, "Error must be nil")
		assert.Equal(t, "Fake Post", input.Name, "Name must be parsed from body")
		assert.Equal(t, fakeDate, input.Date, "Date must be parsed from body")
	})

	t.Run("invalid body input, base64", func(t *testing.T) {
		var input mockPostRequest
		err := UnmarshalRequest(
			events.APIGatewayProxyRequest{
				IsBase64Encoded: true,
				Body:            "dGhpcyBpcyBub3QgSlNPTg==",
			},
			true,
			&input,
		)

		assert.NotEqual(t, nil, err, "Error must not be nil")
	})
}
