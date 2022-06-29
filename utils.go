package lmdrouter

import (
	"github.com/aws/aws-lambda-go/events"
	"math/rand"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func GenerateAPIGatewayContextFromContext() events.APIGatewayProxyRequestContext {
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

func AreSameSecond(one, two *time.Time) bool {
	return one.Year() == two.Year() &&
		one.Month() == two.Month() &&
		one.Day() == two.Day() &&
		one.Hour() == two.Hour() &&
		one.Minute() == two.Minute() &&
		one.Second() == two.Second()
}
