// Package util contains utility functions. Most generally, functions that generate random instances of structs
// for the purposes of testing. This prevents using hard coded values in unit tests which can create copy/paste bugs.
package util

import (
	"cloud.google.com/go/civil"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/seantcanavan/lambda_jwt_router/lcom"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"math/rand"
	"time"
)

// GenerateRandomAPIGatewayProxyRequest returns a random instance of events.APIGatewayProxyRequest instance for testing purposes.
func GenerateRandomAPIGatewayProxyRequest() events.APIGatewayProxyRequest {
	body := struct {
		key   string
		value string
	}{
		key:   "key",
		value: "value",
	}

	jsonBytes, _ := json.Marshal(body)

	return events.APIGatewayProxyRequest{
		Resource:                        GenerateRandomString(10),
		Path:                            GenerateRandomString(10),
		HTTPMethod:                      GenerateRandomString(10),
		Headers:                         map[string]string{"headers": "value"},
		MultiValueHeaders:               map[string][]string{"multiValueHeaders": {"hello there"}},
		QueryStringParameters:           map[string]string{"queryStringParameters": "value"},
		MultiValueQueryStringParameters: map[string][]string{"multiValueQueryStringParameters": {"hello there"}},
		PathParameters:                  map[string]string{"pathParameters": "value"},
		StageVariables:                  map[string]string{"stageVariables": "value"},
		RequestContext:                  GenerateRandomAPIGatewayContext(),
		Body:                            string(jsonBytes),
		IsBase64Encoded:                 false,
	}
}

// GenerateRandomAPIGatewayContext returns a random instance of events.APIGatewayProxyRequestContext for testing purposes.
func GenerateRandomAPIGatewayContext() events.APIGatewayProxyRequestContext {
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

// GenerateRandomInt returns a random integer between N and M (inclusive) for testing purposes.
func GenerateRandomInt(N, M int) int {
	return rand.Intn(M-N+1) + N
}

// GenerateRandomString returns a random string of length N for testing purposes.
func GenerateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

type MockConst string
type Number string
type StringAliasExample string

type MockItem struct {
	ID   string
	Name string
	Date time.Time
}

type MockGetReq struct {
	ID            string `lambda:"path.id"`
	ShowSomething bool   `lambda:"query.show_something"`
}

type MockListReq struct {
	Alias         StringAliasExample    `lambda:"query.alias"`
	AliasPtr      *StringAliasExample   `lambda:"query.alias_ptr"`
	Bool1         bool                  `lambda:"query.bool1"`
	Bool2         bool                  `lambda:"query.bool2"`
	Bool3         bool                  `lambda:"query.bool3"`
	Bool4         bool                  `lambda:"query.bool4"`
	Bool5         bool                  `lambda:"query.bool5"`
	Bool6         bool                  `lambda:"query.bool6"`
	Bool7         bool                  `lambda:"query.bool7"`
	Bool8         bool                  `lambda:"query.bool8"`
	Bool9         bool                  `lambda:"query.bool9"`
	Civil         civil.Date            `lambda:"query.civil"`
	CivilPtr      *civil.Date           `lambda:"query.civilPtr"`
	CivilPtrNil   *civil.Date           `lambda:"query.civilPtrNil"`
	CommaSplit    []Number              `lambda:"query.commaSplit"`
	CommaSplitPtr []*Number             `lambda:"query.commaSplitPtr"`
	Const         MockConst             `lambda:"query.const"`
	ConstPtr      *MockConst            `lambda:"query.constPtr"`
	ConstPtrNil   *MockConst            `lambda:"query.constPtrNil"`
	Encoding      []string              `lambda:"header.Accept-Encoding"`
	ID            string                `lambda:"path.id"`
	IDs           []*string             `lambda:"query.ids"`
	Language      string                `lambda:"header.Accept-Language"`
	MongoID       primitive.ObjectID    `lambda:"query.mongoId"`
	MongoIDPtr    *primitive.ObjectID   `lambda:"query.mongoIdPtr"`
	MongoIDPtrNil *primitive.ObjectID   `lambda:"query.mongoIdPtrNil"`
	MongoIDs      []primitive.ObjectID  `lambda:"query.mongoIds"`
	MongoIDsPtr   []*primitive.ObjectID `lambda:"query.mongoIdsPtr"`
	Number        *float32              `lambda:"query.number"`
	Numbers       []float64             `lambda:"query.numbers"`
	PBoolOne      *bool                 `lambda:"query.pbool1"`
	PBoolTwo      *bool                 `lambda:"query.pbool2"`
	Page          int64                 `lambda:"query.page"`
	PageSize      *int64                `lambda:"query.page_size"`
	Terms         []string              `lambda:"query.terms"`
	Time          time.Time             `lambda:"query.time"`
	TimePtr       *time.Time            `lambda:"query.timePtr"`
	TimePtrNil    *time.Time            `lambda:"query.timePtrNil"`
}

type MockPostReq struct {
	ID   string    `lambda:"path.id"`
	Name string    `json:"name"`
	Date time.Time `json:"date"`
}

// GenerateExpandedMapClaims returns a random
func GenerateExpandedMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		lcom.JWTClaimAudienceKey:  GenerateRandomString(10),
		lcom.JWTClaimEmailKey:     GenerateRandomString(10),
		lcom.JWTClaimExpiresAtKey: time.Now().Add(time.Hour * 30).Unix(),
		lcom.JWTClaimFirstNameKey: GenerateRandomString(10),
		lcom.JWTClaimFullNameKey:  GenerateRandomString(10),
		lcom.JWTClaimIDKey:        GenerateRandomString(10),
		lcom.JWTClaimIssuedAtKey:  time.Now().Unix(),
		lcom.JWTClaimIssuerKey:    GenerateRandomString(10),
		lcom.JWTClaimLevelKey:     GenerateRandomString(10),
		lcom.JWTClaimNotBeforeKey: time.Now().Add(time.Hour * -1).Unix(),
		lcom.JWTClaimSubjectKey:   GenerateRandomString(10),
		lcom.JWTClaimUserTypeKey:  GenerateRandomString(10),
	}
}

func GenerateStandardMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		lcom.JWTClaimAudienceKey:  GenerateRandomString(10),
		lcom.JWTClaimExpiresAtKey: time.Now().Add(time.Hour * 30).Unix(),
		lcom.JWTClaimIDKey:        GenerateRandomString(10),
		lcom.JWTClaimIssuedAtKey:  time.Now().Unix(),
		lcom.JWTClaimIssuerKey:    GenerateRandomString(10),
		lcom.JWTClaimNotBeforeKey: time.Now().Add(time.Hour * -1).Unix(),
		lcom.JWTClaimSubjectKey:   GenerateRandomString(10),
	}
}

func WrapErrors(err1, err2 error) error {
	return fmt.Errorf(err1.Error()+": %w", err2)
}
