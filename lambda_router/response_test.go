package lambda_router

import (
	"encoding/base64"
	"errors"
	"net/http"
	"testing"

	"github.com/jgroeneveld/trial/assert"
)

type customStruct struct {
	StructKey string `json:"structKey"`
}

func TestCustomRes(t *testing.T) {
	httpStatus := http.StatusTeapot
	headers := map[string]string{
		"key": "value",
	}

	structValue := customStruct{
		StructKey: "structValue",
	}

	res, err := CustomRes(httpStatus, headers, structValue)
	assert.Nil(t, err)

	t.Run("verify CustomRes returns the struct in the response body", func(t *testing.T) {
		var returnedStruct customStruct
		err = UnmarshalRes(res, &returnedStruct)
		assert.Nil(t, err)

		assert.Equal(t, structValue, returnedStruct)
	})
	t.Run("verify CustomRes returns the key value pair in the response headers", func(t *testing.T) {
		assert.Equal(t, res.Headers["key"], headers["key"])
	})
	t.Run("verify CustomRes returns the correct status code", func(t *testing.T) {
		assert.Equal(t, httpStatus, res.StatusCode)
	})
}

func TestEmptyRes(t *testing.T) {
	res, err := EmptyRes()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Nil(t, err)
	assert.Equal(t, "{}", res.Body)

	t.Run("verify EmptyRes returns the correct status code", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
}

func TestErrorRes(t *testing.T) {
	t.Run("Handle an StatusAndErrorRes", func(t *testing.T) {
		res, _ := ErrorRes(HTTPError{
			Status:  http.StatusBadRequest,
			Message: "Invalid input",
		})
		assert.Equal(t, http.StatusBadRequest, res.StatusCode, "status status must be correct")
		assert.Equal(t, `{"status":400,"message":"Invalid input"}`, res.Body, "body must be correct")
	})
	t.Run("Handle an StatusAndErrorRes when ExposeServerErrors is true", func(t *testing.T) {
		ExposeServerErrors = true
		res, _ := ErrorRes(HTTPError{
			Status:  http.StatusInternalServerError,
			Message: "database down",
		})
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		assert.Equal(t, `{"status":500,"message":"database down"}`, res.Body, "body must be correct")
	})
	t.Run("Handle an StatusAndErrorRes when ExposeServerErrors is false", func(t *testing.T) {
		ExposeServerErrors = false
		res, _ := ErrorRes(HTTPError{
			Status:  http.StatusInternalServerError,
			Message: "database down",
		})
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		assert.Equal(t, `{"status":500,"message":"Internal Server Error"}`, res.Body, "body must be correct")
	})
	t.Run("Handle a general error when ExposeServerErrors is true", func(t *testing.T) {
		ExposeServerErrors = true
		res, _ := ErrorRes(errors.New("database down"))
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		assert.Equal(t, `{"status":500,"message":"database down"}`, res.Body, "body must be correct")
	})
	t.Run("Handle a general error when ExposeServerErrors is false", func(t *testing.T) {
		ExposeServerErrors = false
		res, _ := ErrorRes(errors.New("database down"))
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		assert.Equal(t, `{"status":500,"message":"Internal Server Error"}`, res.Body, "body must be correct")
	})
}

func TestFileRes(t *testing.T) {
	csvContent := `
header1, header2
value1, value2
`
	res, err := FileRes("text/csv", map[string]string{"key": "value"}, []byte(csvContent))
	assert.Nil(t, err)

	t.Run("verify FileRes returns the correct status code", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("verify FileRes marks the response as NOT base64 encoded", func(t *testing.T) {
		assert.False(t, res.IsBase64Encoded)
	})
	t.Run("verify FileRes embeds the bytes correctly in the response object as a string", func(t *testing.T) {
		assert.Equal(t, csvContent, res.Body)
	})
	t.Run("verify FileRes preserves the original header values", func(t *testing.T) {
		assert.Equal(t, "value", res.Headers["key"])
	})
}

func TestFileB64Res(t *testing.T) {
	csvContent := `
header1, header2
value1, value2
`
	res, err := FileB64Res("text/csv", map[string]string{"key": "value"}, []byte(csvContent))
	assert.Nil(t, err)

	t.Run("verify FileB64Res returns the correct status code", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("verify FileB64Res marks the response as base64 encoded", func(t *testing.T) {
		assert.True(t, res.IsBase64Encoded)
	})
	t.Run("verify FileB64Res embeds the bytes correctly in the response object as a byte64 encoded string", func(t *testing.T) {
		b64CSVContent := res.Body

		decodedCSVContent, decodeErr := base64.StdEncoding.DecodeString(b64CSVContent)
		assert.Nil(t, decodeErr)
		assert.Equal(t, csvContent, string(decodedCSVContent))
	})
	t.Run("verify FileRes preserves the original header values", func(t *testing.T) {
		assert.Equal(t, "value", res.Headers["key"])

	})
}

func TestStatusAndErrorRes(t *testing.T) {
	newErr := errors.New("hello there")
	res, err := StatusAndErrorRes(http.StatusTeapot, newErr)
	assert.Nil(t, err)

	t.Run("verify StatusAndErrorRes returns the correct status code", func(t *testing.T) {
		assert.Equal(t, http.StatusTeapot, res.StatusCode)
	})
}

func TestSuccessRes(t *testing.T) {
	cs := customStruct{StructKey: "hello there"}
	res, err := SuccessRes(cs)
	assert.Nil(t, err)
	t.Run("verify SuccessRes returns the correct status code", func(t *testing.T) {
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("verify SuccessRes returns the struct in the response body", func(t *testing.T) {
		var returnedStruct customStruct
		unmarshalErr := UnmarshalRes(res, &returnedStruct)
		assert.Nil(t, unmarshalErr)
		assert.Equal(t, cs, returnedStruct)
	})
}
