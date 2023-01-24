package utils

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/seantcanavan/lmdrouter"
	"github.com/seantcanavan/lmdrouter/response"
	"math/rand"
	"net/http"
	"time"
)

func GenerateAPIGatewayProxyReq() events.APIGatewayProxyRequestContext {
	return events.APIGatewayProxyRequestContext{
		AccountID:     GenerateRandomString(10),
		ResourceID:    GenerateRandomString(10),
		OperationName: GenerateRandomString(10),
		Stage:         GenerateRandomString(10),
		DomainName:    GenerateRandomString(10),
		DomainPrefix:  GenerateRandomString(10),
		RequestID:     GenerateRandomString(10),
		Protocol:      GenerateRandomString(10),
		Identity: events.APIGatewayRequestIdentity{
			CognitoIdentityPoolID:         GenerateRandomString(10),
			AccountID:                     GenerateRandomString(10),
			CognitoIdentityID:             GenerateRandomString(10),
			Caller:                        GenerateRandomString(10),
			APIKey:                        GenerateRandomString(10),
			APIKeyID:                      GenerateRandomString(10),
			AccessKey:                     GenerateRandomString(10),
			SourceIP:                      GenerateRandomString(10),
			CognitoAuthenticationType:     GenerateRandomString(10),
			CognitoAuthenticationProvider: GenerateRandomString(10),
			UserArn:                       GenerateRandomString(10),
			UserAgent:                     GenerateRandomString(10),
			User:                          GenerateRandomString(10),
		},
		ResourcePath:     GenerateRandomString(10),
		Path:             GenerateRandomString(10),
		Authorizer:       map[string]interface{}{"hi there": "sean"},
		HTTPMethod:       GenerateRandomString(10),
		RequestTime:      GenerateRandomString(10),
		RequestTimeEpoch: 0,
		APIID:            GenerateRandomString(10),
	}
}

func GenerateEmptySuccessHandler() lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return response.Empty()
	}
}

func GenerateEmptyErrorHandler() lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return response.ErrorAndStatus(http.StatusInternalServerError, errors.New("this error is simulated"))
	}
}

func GenerateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
