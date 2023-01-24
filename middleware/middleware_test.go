package middleware

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jgroeneveld/trial/assert"
	"github.com/joho/godotenv"
	"github.com/seantcanavan/lmdrouter/jwt_auth"
	"github.com/seantcanavan/lmdrouter/utils"
	"log"
	"net/http"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
}

func setup() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Unable to load .env file: %s", err)
	}
}

func TestAllowOptions(t *testing.T) {
	t.Run("verify empty OPTIONS req succeeds", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodOptions,
			Headers:        nil,
			RequestContext: utils.GenerateAPIGatewayProxyReq(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptions(utils.GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})
	t.Run("verify OPTIONS req succeeds with invalid JWT for AllowOptions", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		jwt, err := jwt_auth.Sign(nil)
		assert.Nil(t, err)

		jwt = jwt + "hi" // create an invalid JWT

		req := events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodOptions,
			Headers: map[string]string{
				"Authorization": "Bearer " + jwt,
			},
			RequestContext: utils.GenerateAPIGatewayProxyReq(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptions(utils.GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})
	t.Run("verify OPTIONS req succeeds with no Authorization header for AllowOptions", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		req := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodOptions,
			Headers:        nil,
			RequestContext: utils.GenerateAPIGatewayProxyReq(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptions(utils.GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})
}

//
//func TestDecodeJWTMiddleware(t *testing.T) {
//	t.Run("successful jwt sign and set context adminFullName and adminID and adminLevel with GET req", func(t *testing.T) {
//		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
//		defer cancel()
//
//		audience := utils.GenerateRandomString(10)
//		id := utils.GenerateRandomString(10)
//		issuer := utils.GenerateRandomString(10)
//		subject := utils.GenerateRandomString(10)
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
//			RequestContext: GenerateAPIGatewayProxyReq(),
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

//}
//

//
//func AreSameSecond(one, two *time.Time) bool {
//	return one.Year() == two.Year() &&
//		one.Month() == two.Month() &&
//		one.Day() == two.Day() &&
//		one.Hour() == two.Hour() &&
//		one.Minute() == two.Minute() &&
//		one.Second() == two.Second()
//}
