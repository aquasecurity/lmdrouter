package lambda_router

import (
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jgroeneveld/trial/assert"
)

type mockConst string

const (
	mockConstTwo mockConst = "two"
)

type Number string

const (
	numberOne   Number = "one"
	numberTwo   Number = "two"
	numberThree Number = "three"
)

type mockItem struct {
	ID   string
	Name string
	Date time.Time
}

type mockGetReq struct {
	ID            string `lambda:"path.id"`
	ShowSomething bool   `lambda:"query.show_something"`
}

type mockListReq struct {
	ID         string              `lambda:"path.id"`
	Page       int64               `lambda:"query.page"`
	PageSize   int64               `lambda:"query.page_size"`
	Terms      []string            `lambda:"query.terms"`
	Numbers    []float64           `lambda:"query.numbers"`
	Const      mockConst           `lambda:"query.const"`
	Bool       bool                `lambda:"query.bool"`
	PBoolOne   *bool               `lambda:"query.pbool1"`
	PBoolTwo   *bool               `lambda:"query.pbool2"`
	Time       *time.Time          `lambda:"query.time"`
	Alias      stringAliasExample  `lambda:"query.alias"`
	AliasPtr   *stringAliasExample `lambda:"query.alias_ptr"`
	CommaSplit []Number            `lambda:"query.commaSplit"`
	Language   string              `lambda:"header.Accept-Language"`
	Encoding   []string            `lambda:"header.Accept-Encoding"`
}

type mockPostReq struct {
	ID   string    `lambda:"path.id"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

type stringAliasExample string

const aliasExample stringAliasExample = "world"

func Test_UnmarshalReq(t *testing.T) {
	t.Run("valid path&query input", func(t *testing.T) {
		var input mockListReq
		err := UnmarshalRequest(
			events.APIGatewayProxyRequest{
				PathParameters: map[string]string{
					"id": "fake-scan-id",
				},
				QueryStringParameters: map[string]string{
					"page":       "2",
					"page_size":  "30",
					"const":      "two",
					"bool":       "true",
					"pbool1":     "0",
					"time":       "2021-11-01T11:11:11.000Z",
					"alias":      "hello",
					"alias_ptr":  "world",
					"commaSplit": "one,two,three",
				},
				MultiValueQueryStringParameters: map[string][]string{
					"terms":   {"one", "two"},
					"numbers": {"1.2", "3.5", "666.666"},
				},
				Headers: map[string]string{
					"Accept-Language": "en-us",
				},
				MultiValueHeaders: map[string][]string{
					"Accept-Encoding": {"gzip", "deflate"},
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
		assert.True(t, input.Bool, "Bool must be true")
		assert.NotNil(t, input.PBoolOne, "PBoolOne must not be nil")
		assert.False(t, *input.PBoolOne, "PBoolOne must be *false")
		assert.NotNil(t, input.Time, "Time must not be nil")
		assert.Equal(t, input.Time.Format(time.RFC3339), "2021-11-01T11:11:11Z")
		assert.Equal(t, input.Alias, stringAliasExample("hello"))
		assert.NotNil(t, input.AliasPtr)
		assert.Equal(t, *input.AliasPtr, aliasExample)
		assert.DeepEqual(t, []Number{numberOne, numberTwo, numberThree}, input.CommaSplit, "CommaSplit must have 2 items")
		assert.Equal(t, (*bool)(nil), input.PBoolTwo, "PBoolTwo must be nil")
		assert.DeepEqual(t, []string{"one", "two"}, input.Terms, "Terms must be parsed from multiple query params")
		assert.DeepEqual(t, []float64{1.2, 3.5, 666.666}, input.Numbers, "Numbers must be parsed from multiple query params")
		assert.DeepEqual(t, []string{"gzip", "deflate"}, input.Encoding, "Encoding must be parsed from multiple header params")
	})

	t.Run("invalid path&query input", func(t *testing.T) {
		var input mockListReq
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
		assert.True(t, ok, "Error must be an response.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Status, "Error code must be 400")
	})

	fakeDate := time.Date(2020, 3, 23, 11, 33, 0, 0, time.UTC)

	t.Run("valid body input, not base64", func(t *testing.T) {
		var input mockPostReq
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
		var input mockPostReq
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
		var input mockPostReq
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
		var input mockPostReq
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
