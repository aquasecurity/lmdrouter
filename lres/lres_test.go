package lres

import (
	"encoding/base64"
	"errors"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/seantcanavan/lambda_jwt_router/lcom"
	"github.com/stretchr/testify/require"
	"net/http"
	"os"
	"testing"
)

type customStruct struct {
	StructKey string `json:"structKey"`
}

func TestCustomRes(t *testing.T) {
	headersHeaderVal := util.GenerateRandomString(10)
	methodsHeaderVal := util.GenerateRandomString(10)
	originHeaderVal := util.GenerateRandomString(10)

	setCors(t, headersHeaderVal, methodsHeaderVal, originHeaderVal)
	defer unsetCors(t)

	httpStatus := http.StatusTeapot
	headers := map[string]string{
		"key": "value",
	}

	structValue := customStruct{
		StructKey: "structValue",
	}

	res, err := CustomRes(httpStatus, headers, structValue)
	require.Nil(t, err)

	t.Run("verify CustomRes returns the struct in the response body", func(t *testing.T) {
		var returnedStruct customStruct
		err = UnmarshalRes(res, &returnedStruct)
		require.Nil(t, err)

		require.Equal(t, structValue, returnedStruct)
	})
	t.Run("verify CustomRes returns the key value pair in the response headers", func(t *testing.T) {
		require.Equal(t, res.Headers["key"], headers["key"])
	})
	t.Run("verify CustomRes returns the correct status code", func(t *testing.T) {
		require.Equal(t, httpStatus, res.StatusCode)
	})
	t.Run("verify CustomRes returns CORS headers", func(t *testing.T) {
		require.Equal(t, headersHeaderVal, res.Headers[lcom.CORSHeadersHeaderKey])
		require.Equal(t, methodsHeaderVal, res.Headers[lcom.CORSMethodsHeaderKey])
		require.Equal(t, originHeaderVal, res.Headers[lcom.CORSOriginHeaderKey])
	})
}

func TestEmptyRes(t *testing.T) {
	res, err := EmptyRes()
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Nil(t, err)
	require.Equal(t, "{}", res.Body)

	t.Run("verify EmptyRes returns the correct status code", func(t *testing.T) {
		require.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("verify EmptyRes returns CORS headers", func(t *testing.T) {
		require.Equal(t, res.Headers[lcom.CORSHeadersEnvKey], "")
		require.Equal(t, res.Headers[lcom.CORSMethodsEnvKey], "")
		require.Equal(t, res.Headers[lcom.CORSOriginEnvKey], "")
	})
}

func TestErrorRes(t *testing.T) {
	setCors(t, "*", "*", "*")
	defer unsetCors(t)

	t.Run("Handle an HTTPError ErrorRes without ExposeServerErrors set and verify CORS", func(t *testing.T) {
		res, _ := ErrorRes(HTTPError{
			Status:  http.StatusBadRequest,
			Message: "Invalid input",
		})
		require.Equal(t, http.StatusBadRequest, res.StatusCode, "status status must be correct")
		require.Equal(t, `{"status":400,"message":"Invalid input"}`, res.Body, "body must be correct")
		t.Run("verify ErrorRes returns CORS headers", func(t *testing.T) {
			require.Equal(t, res.Headers[lcom.CORSHeadersHeaderKey], "*")
			require.Equal(t, res.Headers[lcom.CORSMethodsHeaderKey], "*")
			require.Equal(t, res.Headers[lcom.CORSOriginHeaderKey], "*")
		})
	})
	t.Run("Handle an HTTPError for ErrorRes when ExposeServerErrors is true", func(t *testing.T) {
		ExposeServerErrors = true
		res, _ := ErrorRes(HTTPError{
			Status:  http.StatusInternalServerError,
			Message: "database down",
		})
		require.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		require.Equal(t, `{"status":500,"message":"database down"}`, res.Body, "body must be correct")
	})
	t.Run("Handle an HTTPError for ErrorRes when ExposeServerErrors is false", func(t *testing.T) {
		ExposeServerErrors = false
		res, _ := ErrorRes(HTTPError{
			Status:  http.StatusInternalServerError,
			Message: "database down",
		})
		require.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		require.Equal(t, `{"status":500,"message":"Internal Server Error"}`, res.Body, "body must be correct")
	})
	t.Run("Handle a general error for ErrorRes when ExposeServerErrors is true", func(t *testing.T) {
		ExposeServerErrors = true
		res, _ := ErrorRes(errors.New("database down"))
		require.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		require.Equal(t, `{"status":500,"message":"database down"}`, res.Body, "body must be correct")
	})
	t.Run("Handle a general error for ErrorRes when ExposeServerErrors is false", func(t *testing.T) {
		ExposeServerErrors = false
		res, _ := ErrorRes(errors.New("database down"))
		require.Equal(t, http.StatusInternalServerError, res.StatusCode, "status must be correct")
		require.Equal(t, `{"status":500,"message":"Internal Server Error"}`, res.Body, "body must be correct")
	})
}

func TestFileRes(t *testing.T) {
	setCors(t, "*", "*", "*")
	defer unsetCors(t)

	csvContent := `
header1, header2
value1, value2
`
	res, err := FileRes("text/csv", map[string]string{"key": "value"}, []byte(csvContent))
	require.Nil(t, err)

	t.Run("verify FileRes returns the correct status code", func(t *testing.T) {
		require.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("verify FileRes marks the response as NOT base64 encoded", func(t *testing.T) {
		require.False(t, res.IsBase64Encoded)
	})
	t.Run("verify FileRes embeds the bytes correctly in the response object as a string", func(t *testing.T) {
		require.Equal(t, csvContent, res.Body)
	})
	t.Run("verify FileRes preserves the original header values", func(t *testing.T) {
		require.Equal(t, "value", res.Headers["key"])
	})
	t.Run("verify FileRes returns CORS headers", func(t *testing.T) {
		require.Equal(t, res.Headers[lcom.CORSHeadersHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSMethodsHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSOriginHeaderKey], "*")
	})
}

func TestFileB64Res(t *testing.T) {
	setCors(t, "*", "*", "*")
	defer unsetCors(t)

	csvContent := `
header1, header2
value1, value2
`
	res, err := FileB64Res("text/csv", map[string]string{"key": "value"}, []byte(csvContent))
	require.Nil(t, err)

	t.Run("verify FileB64Res returns the correct status code", func(t *testing.T) {
		require.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("verify FileB64Res marks the response as base64 encoded", func(t *testing.T) {
		require.True(t, res.IsBase64Encoded)
	})
	t.Run("verify FileB64Res embeds the bytes correctly in the response object as a byte64 encoded string", func(t *testing.T) {
		b64CSVContent := res.Body

		decodedCSVContent, decodeErr := base64.StdEncoding.DecodeString(b64CSVContent)
		require.Nil(t, decodeErr)
		require.Equal(t, csvContent, string(decodedCSVContent))
	})
	t.Run("verify FileRes preserves the original header values", func(t *testing.T) {
		require.Equal(t, "value", res.Headers["key"])
	})
	t.Run("verify FileB64Res returns CORS headers", func(t *testing.T) {
		require.Equal(t, res.Headers[lcom.CORSHeadersHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSMethodsHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSOriginHeaderKey], "*")
	})
}

func TestStatusAndErrorRes(t *testing.T) {
	setCors(t, "*", "*", "*")
	defer unsetCors(t)

	newErr := errors.New("hello there")
	res, err := StatusAndErrorRes(http.StatusTeapot, newErr)
	require.Nil(t, err)

	t.Run("verify StatusAndErrorRes returns the correct status code", func(t *testing.T) {
		require.Equal(t, http.StatusTeapot, res.StatusCode)
	})
	t.Run("verify StatusAndErrorRes returns CORS headers", func(t *testing.T) {
		require.Equal(t, res.Headers[lcom.CORSHeadersHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSMethodsHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSOriginHeaderKey], "*")
	})

}

func TestSuccessRes(t *testing.T) {
	setCors(t, "*", "*", "*")
	defer unsetCors(t)

	cs := customStruct{StructKey: "hello there"}
	res, err := SuccessRes(cs)
	require.Nil(t, err)
	t.Run("verify SuccessRes returns the correct status code", func(t *testing.T) {
		require.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("verify SuccessRes returns the struct in the response body", func(t *testing.T) {
		var returnedStruct customStruct
		unmarshalErr := UnmarshalRes(res, &returnedStruct)
		require.Nil(t, unmarshalErr)
		require.Equal(t, cs, returnedStruct)
	})
	t.Run("verify SuccessRes returns CORS headers", func(t *testing.T) {
		require.Equal(t, res.Headers[lcom.CORSHeadersHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSMethodsHeaderKey], "*")
		require.Equal(t, res.Headers[lcom.CORSOriginHeaderKey], "*")
	})
}

func setCors(t *testing.T, headers, methods, origin string) {
	err := os.Setenv(lcom.CORSHeadersEnvKey, headers)
	require.NoError(t, err)

	err = os.Setenv(lcom.CORSMethodsEnvKey, methods)
	require.NoError(t, err)

	err = os.Setenv(lcom.CORSOriginEnvKey, origin)
	require.NoError(t, err)
}

func unsetCors(t *testing.T) {
	err := os.Unsetenv(lcom.CORSHeadersEnvKey)
	require.NoError(t, err)

	err = os.Unsetenv(lcom.CORSMethodsEnvKey)
	require.NoError(t, err)

	err = os.Unsetenv(lcom.CORSOriginEnvKey)
	require.NoError(t, err)
}
