package response

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

const ContentTypeKey = "Content-Type"

// ExposeServerErrors is a boolean indicating whether the Error function
// should expose errors of status code 500 or above to clients. If false, the
// name of the status code is used as the error message instead.
var ExposeServerErrors = true

// Success wraps Custom assuming a 200 OK status code and no
// custom headers to return. This was such a common use case I felt it
// necessary to create a wrapper to make everyone's life easier.
func Success(data interface{}) (events.APIGatewayProxyResponse, error) {
	return Custom(http.StatusOK, nil, data)
}

// Empty returns a simple empty events.APIGatewayProxyResponse with http.StatusOK
func Empty() (events.APIGatewayProxyResponse, error) {
	return Custom(http.StatusOK, nil, struct{}{})
}

// Custom generated an events.APIGatewayProxyResponse object that can
// be directly returned via the lambda's handler function. It receives an HTTP
// status code for the response, a map of HTTP headers (can be empty or nil),
// and a value (probably a struct) representing the response body. This value
// will be marshaled to JSON (currently without base 64 encoding).
func Custom(httpStatus int, headers map[string]string, data interface{}) (
	events.APIGatewayProxyResponse,
	error,
) {

	b, err := json.Marshal(data)
	if err != nil {
		httpStatus = http.StatusInternalServerError
		b = []byte(`{"code":500,"message":"the server has encountered an unexpected error"}`)
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	headers["Content-Type"] = "application/json; charset=UTF-8"

	return events.APIGatewayProxyResponse{
		StatusCode:      httpStatus,
		IsBase64Encoded: false,
		Headers:         addCors(headers),
		Body:            string(b),
	}, nil
}

// File generates a new events.APIGatewayProxyResponse with the ContentTypeKey header set appropriately, the
// file bytes encoded to base64, and the http status set to http.StatusOK
func File(contentType string, headers map[string]string, fileBytes []byte) (events.APIGatewayProxyResponse, error) {
	if headers == nil {
		headers = map[string]string{
			ContentTypeKey: contentType,
		}
	} else {
		headers[ContentTypeKey] = contentType
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		Headers:         addCors(headers),
		Body:            base64.StdEncoding.EncodeToString(fileBytes),
		IsBase64Encoded: true,
	}, nil
}

// Error generates an events.APIGatewayProxyResponse from an error value.
// If the error is an HTTPError, the response's status code will be taken from
// the error. Otherwise, the error is assumed to be 500 Internal Server Error.
// Regardless, all errors will generate a JSON response in the format
// `{ "code": 500, "error": "something failed" }`
// This format cannot currently be changed. If you do not wish to expose server
// errors (i.e. errors whose status code is 500 or above), set the
// ExposeServerErrors global variable to false.
func Error(err error) (events.APIGatewayProxyResponse, error) {
	var httpErr HTTPError
	if !errors.As(err, &httpErr) {
		httpErr = HTTPError{
			Status:  http.StatusInternalServerError,
			Message: err.Error(),
		}
	}

	if httpErr.Status >= 500 && !ExposeServerErrors {
		httpErr.Message = http.StatusText(httpErr.Status)
	}

	return Custom(httpErr.Status, nil, httpErr)
}

// ErrorAndStatus generates a custom error return response with the given http status code and error.
// Setting ExposeServerErrors to false will prevent leaking data to clients.
func ErrorAndStatus(httpStatus int, err error) (events.APIGatewayProxyResponse, error) {
	httpErr := HTTPError{
		Status:  httpStatus,
		Message: err.Error(),
	}

	// If we're not exposing server errors then return a general message
	if httpErr.Status >= 500 && !ExposeServerErrors {
		httpErr.Message = http.StatusText(httpErr.Status)
	}

	return Custom(httpErr.Status, nil, httpErr)
}

// addCors injects CORS Origin and CORS Methods headers into the response object before it's returned.
func addCors(headers map[string]string) map[string]string {
	corsMethods := os.Getenv("CORS_METHODS")
	corsOrigins := os.Getenv("CORS_ORIGIN")

	if corsMethods == "" {
		corsMethods = "*"
	}

	if corsOrigins == "" {
		corsOrigins = "*"
	}

	headers["Access-Control-Allow-Headers"] = "*"
	headers["Access-Control-Allow-Methods"] = corsMethods
	headers["Access-Control-Allow-Origin"] = corsOrigins

	return headers
}

// HTTPError is a generic struct type for JSON error responses. It allows the library
// to assign an HTTP status code for the errors returned by its various functions.
type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// Error returns a string representation of the HTTPError instance.
func (err HTTPError) Error() string {
	return fmt.Sprintf("error %d: %s", err.Status, err.Message)
}
