package lmdrouter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// MarshalResponse generated an events.APIGatewayProxyResponse object that can
// be directly returned via the lambda's handler function. It receives an HTTP
// status code for the response, a map of HTTP headers (can be empty or nil),
// and a value (probably a struct) representing the response body. This value
// will be marshaled to JSON (currently without base 64 encoding).
func MarshalResponse(status int, headers map[string]string, data interface{}) (
	events.APIGatewayProxyResponse,
	error,
) {
	b, err := json.Marshal(data)
	if err != nil {
		status = http.StatusInternalServerError
		b = []byte(`{"code":500,"message":"the server has encountered an unexpected error"}`)
	}

	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json; charset=UTF-8"

	return events.APIGatewayProxyResponse{
		StatusCode:      status,
		IsBase64Encoded: false,
		Headers:         headers,
		Body:            string(b),
	}, nil
}

// ExposeServerErrors is a boolean indicating whether the HandleError function
// should expose errors of status code 500 or above to clients. If false, the
// name of the status code is used as the error message instead.
var ExposeServerErrors = true

// HandleError generates an events.APIGatewayProxyResponse from an error value.
// If the error is an HTTPError, the response's status code will be taken from
// the error. Otherwise, the error is assumed to be 500 Internal Server Error.
// Regardless, all errors will generate a JSON response in the format
// `{ "code": 500, "error": "something failed" }`
// This format cannot currently be changed. If you do not wish to expose server
// errors (i.e. errors whose status code is 500 or above), set the
// ExposeServerErrors global variable to false.
func HandleError(err error) (events.APIGatewayProxyResponse, error) {
	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		httpErr = HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	if httpErr.Code >= 500 && !ExposeServerErrors {
		httpErr.Message = http.StatusText(httpErr.Code)
	}

	return MarshalResponse(httpErr.Code, nil, httpErr)
}
