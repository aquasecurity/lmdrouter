package lmdrouter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

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

func HandleError(err error) (events.APIGatewayProxyResponse, error) {
	var httpErr HTTPError
	if errors.As(err, &httpErr) {
		return MarshalResponse(httpErr.Code, nil, httpErr)
	}

	return MarshalResponse(
		http.StatusInternalServerError,
		nil,
		HTTPError{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
		},
	)
}
