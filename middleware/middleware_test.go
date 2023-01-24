package middleware

//
//import (
//	"context"
//	"encoding/json"
//	"github.com/aws/aws-lambda-go/events"
//	"github.com/jgroeneveld/trial/assert"
//	"net/http"
//	"strings"
//	"testing"
//	"time"
//)
//
//func getEmptyHandler() Handler {
//	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
//		events.APIGatewayProxyResponse,
//		error) {
//		return MarshalResponse(200,
//			map[string]string{
//				AudienceKey: ctx.Value(AudienceKey).(string),
//				IDKey:       ctx.Value(IDKey).(string),
//				IssuerKey:   ctx.Value(IssuerKey).(string),
//				SubjectKey:  ctx.Value(SubjectKey).(string),
//			}, nil)
//	}
//}
//
//func TestDecodeJWTMiddleware(t *testing.T) {
//	t.Run("missing Authorization header", func(t *testing.T) {
//		req := events.APIGatewayProxyRequest{}
//		jwtMiddlewareHandler := DecodeStandardJWTMiddleware(getEmptyHandler())
//		res, err := jwtMiddlewareHandler(nil, req)
//		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
//		assert.Equal(t, res.StatusCode, http.StatusUnauthorized)
//
//		var responseBody HTTPError
//		err = json.Unmarshal([]byte(res.Body), &responseBody)
//		assert.Nil(t, err)
//
//		assert.Equal(t, responseBody.Status, res.StatusCode)
//		assert.Equal(t, responseBody.Message, "missing Authorization header value")
//	})
//
//	t.Run("successful jwt sign and set context adminFullName and adminID and adminLevel with GET req", func(t *testing.T) {
//		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
//		defer cancel()
//
//		audience := GenerateRandomString(10)
//		id := GenerateRandomString(10)
//		issuer := GenerateRandomString(10)
//		subject := GenerateRandomString(10)
//
//		claims, err := InitializeMapClaims(
//			audience,
//			time.Now().Add(time.Hour*1).Unix(), // expiresAt
//			id,
//			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
//			issuer,
//			time.Now().Unix()-50000, // notBefore
//			subject,
//		)
//		assert.Nil(t, err)
//
//		jwt, httpStatus, err := GenerateJWT(claims)
//		assert.Equal(t, httpStatus, http.StatusOK)
//		assert.Nil(t, err)
//
//		headers := map[string]string{
//			"Authorization": "Bearer " + jwt,
//		}
//
//		req := events.APIGatewayProxyRequest{
//			HTTPMethod:     "GET",
//			Headers:        headers,
//			RequestContext: GenerateAPIGatewayContextFromContext(),
//		}
//
//		jwtMiddlewareHandler := DecodeStandardJWTMiddleware(getEmptyHandler())
//		res, err := jwtMiddlewareHandler(ctx, req)
//		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
//		assert.Equal(t, res.StatusCode, http.StatusOK)
//		assert.Equal(t, res.Headers[AudienceKey], audience)
//		assert.Equal(t, res.Headers[IDKey], id)
//		assert.Equal(t, res.Headers[IssuerKey], issuer)
//		assert.Equal(t, res.Headers[SubjectKey], subject)
//		// these headers aren't used in actual responses - they're here to
//		//show the values are correctly injected from context during runtime
//	})
//
//	t.Run("OPTIONS req succeeds with invalid JWT", func(t *testing.T) {
//		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
//		defer cancel()
//
//		jwt, httpStatus, err := GenerateJWT(nil)
//		assert.Equal(t, httpStatus, http.StatusOK)
//		assert.Nil(t, err)
//
//		jwt = jwt + "hi" // create an invalid JWT
//
//		headers := make(map[string]string)
//		headers["Authorization"] = "Bearer " + jwt
//
//		req := events.APIGatewayProxyRequest{
//			HTTPMethod:     "OPTIONS",
//			Headers:        headers,
//			RequestContext: GenerateAPIGatewayContextFromContext(),
//		}
//
//		jwtMiddlewareHandler := DecodeStandardJWTMiddleware(getEmptyHandler())
//		res, err := jwtMiddlewareHandler(ctx, req)
//		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
//		assert.Equal(t, res.StatusCode, http.StatusOK)
//	})
//
//	t.Run("OPTIONS req succeeds with no Authorization header", func(t *testing.T) {
//		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
//		defer cancel()
//
//		req := events.APIGatewayProxyRequest{
//			HTTPMethod:     "OPTIONS",
//			Headers:        nil,
//			RequestContext: GenerateAPIGatewayContextFromContext(),
//		}
//
//		jwtMiddlewareHandler := DecodeStandardJWTMiddleware(getEmptyHandler())
//		res, err := jwtMiddlewareHandler(ctx, req)
//		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
//		assert.Equal(t, res.StatusCode, http.StatusOK)
//	})
//}
//

//
//func GenerateAPIGatewayContextFromContext() events.APIGatewayProxyRequestContext {
//	return events.APIGatewayProxyRequestContext{
//		AccountID:     GenerateRandomString(10),
//		ResourceID:    GenerateRandomString(10),
//		OperationName: GenerateRandomString(10),
//		Stage:         GenerateRandomString(10),
//		DomainName:    GenerateRandomString(10),
//		DomainPrefix:  GenerateRandomString(10),
//		RequestID:     GenerateRandomString(10),
//		Protocol:      GenerateRandomString(10),
//		Identity: events.APIGatewayRequestIdentity{
//			CognitoIdentityPoolID:         GenerateRandomString(10),
//			AccountID:                     GenerateRandomString(10),
//			CognitoIdentityID:             GenerateRandomString(10),
//			Caller:                        GenerateRandomString(10),
//			APIKey:                        GenerateRandomString(10),
//			APIKeyID:                      GenerateRandomString(10),
//			AccessKey:                     GenerateRandomString(10),
//			SourceIP:                      GenerateRandomString(10),
//			CognitoAuthenticationType:     GenerateRandomString(10),
//			CognitoAuthenticationProvider: GenerateRandomString(10),
//			UserArn:                       GenerateRandomString(10),
//			UserAgent:                     GenerateRandomString(10),
//			User:                          GenerateRandomString(10),
//		},
//		ResourcePath:     GenerateRandomString(10),
//		Path:             GenerateRandomString(10),
//		Authorizer:       map[string]interface{}{"hi there": "sean"},
//		HTTPMethod:       GenerateRandomString(10),
//		RequestTime:      GenerateRandomString(10),
//		RequestTimeEpoch: 0,
//		APIID:            GenerateRandomString(10),
//	}
//}
//
//func AreSameSecond(one, two *time.Time) bool {
//	return one.Year() == two.Year() &&
//		one.Month() == two.Month() &&
//		one.Day() == two.Day() &&
//		one.Hour() == two.Hour() &&
//		one.Minute() == two.Minute() &&
//		one.Second() == two.Second()
//}
