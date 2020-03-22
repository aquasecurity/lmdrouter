package lmdrouter

import (
	"errors"
	"net/http"
	"testing"

	"github.com/jgroeneveld/trial/assert"
)

func Test_UnmarshalRequest(t *testing.T) {
	t.Run("valid input", func(t *testing.T) {
		var input mockListRequest
		err := UnmarshalRequest(
			map[string]string{
				"id": "fake-scan-id",
			},
			map[string]string{
				"page":      "2",
				"page_size": "30",
			},
			&input,
		)
		assert.Equal(t, nil, err, "Error must be nil")
		assert.Equal(t, "fake-scan-id", input.ID, "ID must be parsed from path")
		assert.Equal(t, int64(2), input.Page, "Page must be parsed from query")
		assert.Equal(t, int64(30), input.PageSize, "PageSize must be parsed from query")
	})

	t.Run("invalid input", func(t *testing.T) {
		var input mockListRequest
		err := UnmarshalRequest(
			map[string]string{
				"id": "fake-scan-id",
			},
			map[string]string{
				"page": "abcd",
			},
			&input,
		)
		assert.NotEqual(t, nil, err, "Error must not be nil")
		var httpErr HTTPError
		ok := errors.As(err, &httpErr)
		assert.True(t, ok, "Error must be an HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code, "Error code must be 400")
	})
}
