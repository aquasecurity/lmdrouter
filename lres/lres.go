package lres

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/seantcanavan/lambda_jwt_router/lcom"
	"net/http"
	"os"
	"reflect"
)

// ExposeServerErrors is a boolean indicating whether the ErrorRes function
// should expose errors of status code 500 or above to clients. If false, the
// name of the status code is used as the error message instead.
var ExposeServerErrors = true

// CustomRes generated an events.APIGatewayProxyResponse object that can
// be directly returned via the lambda's handler function. It receives an HTTP
// status code for the response, a map of HTTP headers (can be empty or nil),
// and a value (probably a struct) representing the response body. This value
// will be marshaled to JSON (currently without base 64 encoding).
func CustomRes(httpStatus int, headers map[string]string, data interface{}) (
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

	headers[lcom.ContentTypeKey] = "application/json; charset=UTF-8"

	return events.APIGatewayProxyResponse{
		StatusCode:      httpStatus,
		IsBase64Encoded: false,
		Headers:         addCors(headers),
		Body:            string(b),
	}, nil
}

// EmptyRes returns a simple empty events.APIGatewayProxyResponse with http.StatusOK
func EmptyRes() (events.APIGatewayProxyResponse, error) {
	return CustomRes(http.StatusOK, nil, struct{}{})
}

// ErrorRes generates an events.APIGatewayProxyResponse from an error value.
// If the error is an HTTPError, the response's status code will be taken from
// the error. Otherwise, the error is assumed to be 500 Internal Server Error.
// Regardless, all errors will generate a JSON response in the format
// `{ "code": 500, "error": "something failed" }`
// This format cannot currently be changed. If you do not wish to expose server
// errors (i.e. errors whose status code is 500 or above), set the
// ExposeServerErrors global variable to false.
func ErrorRes(err error) (events.APIGatewayProxyResponse, error) {
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

	return CustomRes(httpErr.Status, nil, httpErr)
}

// FileRes generates a new events.APIGatewayProxyResponse with the ContentTypeKey header set appropriately, the
// file bytes added to the response body, and the http status set to http.StatusOK
func FileRes(contentType string, headers map[string]string, fileBytes []byte) (events.APIGatewayProxyResponse, error) {
	if headers == nil {
		headers = map[string]string{
			lcom.ContentTypeKey: contentType,
		}
	} else {
		headers[lcom.ContentTypeKey] = contentType
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		Headers:         addCors(headers),
		Body:            string(fileBytes),
		IsBase64Encoded: false,
	}, nil
}

// FileB64Res generates a new events.APIGatewayProxyResponse with the ContentTypeKey header set appropriately, the
// file bytes encoded to base64, and the http status set to http.StatusOK
func FileB64Res(contentType string, headers map[string]string, fileBytes []byte) (events.APIGatewayProxyResponse, error) {
	if headers == nil {
		headers = map[string]string{
			lcom.ContentTypeKey: contentType,
		}
	} else {
		headers[lcom.ContentTypeKey] = contentType
	}

	return events.APIGatewayProxyResponse{
		StatusCode:      http.StatusOK,
		Headers:         addCors(headers),
		Body:            base64.StdEncoding.EncodeToString(fileBytes),
		IsBase64Encoded: true,
	}, nil
}

// StatusAndErrorRes generates a custom error return response with the given http status code and error.
// Setting ExposeServerErrors to false will prevent leaking data to clients.
func StatusAndErrorRes(httpStatus int, err error) (events.APIGatewayProxyResponse, error) {
	httpErr := HTTPError{
		Status:  httpStatus,
		Message: err.Error(),
	}

	// If we're not exposing server errors then return a general message
	if httpErr.Status >= 500 && !ExposeServerErrors {
		httpErr.Message = http.StatusText(httpErr.Status)
	}

	return CustomRes(httpErr.Status, nil, httpErr)
}

// SuccessRes wraps CustomRes assuming a http.StatusOK status code and no
// custom headers to return. This was such a common use case I felt it
// necessary to create a wrapper to make everyone's life easier.
func SuccessRes(data interface{}) (events.APIGatewayProxyResponse, error) {
	return CustomRes(http.StatusOK, nil, data)
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

// addCors injects CORS Origin and CORS Methods headers into the response object before it's returned.
func addCors(headers map[string]string) map[string]string {
	corsHeaders := os.Getenv(lcom.CORSHeadersEnvKey)
	corsMethods := os.Getenv(lcom.CORSMethodsEnvKey)
	corsOrigins := os.Getenv(lcom.CORSOriginEnvKey)

	if corsHeaders != "" {
		headers[lcom.CORSHeadersHeaderKey] = corsHeaders
	}

	if corsMethods != "" {
		headers[lcom.CORSMethodsHeaderKey] = corsMethods
	}

	if corsOrigins != "" {
		headers[lcom.CORSOriginHeaderKey] = corsOrigins
	}

	return headers
}

// UnmarshalRes should generally be used only when testing as normally you return the response
// directly to the caller and won't need to Unmarshal it. However, if you are testing locally then
// it will help you extract the response body of a lambda request and marshal it to an object.
func UnmarshalRes(res events.APIGatewayProxyResponse, target interface{}) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("invalid unmarshal target, must be pointer to struct")
	}

	return json.Unmarshal([]byte(res.Body), target)
}

// GenerateSuccessHandlerAndMapStandardContext returns a middleware handler
// that takes the values inserted into the context object by DecodeStandardMW
// and returns them as an object from the request so that unit tests can analyze the values
// and make sure they have done the full trip from JWT -> CTX -> unit test
func GenerateSuccessHandlerAndMapStandardContext() lcom.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return CustomRes(http.StatusOK, nil, jwt.StandardClaims{
			Audience:  ctx.Value(lcom.JWTClaimAudienceKey).(string),
			ExpiresAt: ctx.Value(lcom.JWTClaimExpiresAtKey).(int64),
			Id:        ctx.Value(lcom.JWTClaimIDKey).(string),
			IssuedAt:  ctx.Value(lcom.JWTClaimIssuedAtKey).(int64),
			Issuer:    ctx.Value(lcom.JWTClaimIssuerKey).(string),
			NotBefore: ctx.Value(lcom.JWTClaimNotBeforeKey).(int64),
			Subject:   ctx.Value(lcom.JWTClaimSubjectKey).(string),
		})
	}
}
