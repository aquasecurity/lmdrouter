package lmdrouter

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-lambda-go/events"
	"github.com/jgroeneveld/trial/assert"
	"github.com/learnfully/learnfully/lfutils"
	"net/http"
	"testing"
	"time"
)

func getEmptyHandler() Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return MarshalResponse(200, map[string]string{"adminFullName": ctx.Value("adminFullName").(string), "adminID": ctx.Value("adminID").(string), "adminLevel": ctx.Value("adminLevel").(string), "expirationDate": ctx.Value("expirationDate").(time.Time).Format(time.RFC3339)}, nil)
	}
}

func TestDecodeJWTMiddleware(t *testing.T) {
	t.Run("missing Authorization header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		jwtMiddlewareHandler := DecodeJWTMiddleware(getEmptyHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
		assert.Equal(t, res.StatusCode, http.StatusForbidden)

		var responseBody HTTPError
		err = json.Unmarshal([]byte(res.Body), &responseBody)
		assert.Nil(t, err)

		assert.Equal(t, responseBody.Status, res.StatusCode)
		assert.Equal(t, responseBody.Message, "missing Authorization header value")
	})

	t.Run("successful jwt sign and set context adminFullName and adminID and adminLevel with GET req", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		adminFullName := lfutils.GenerateRandomString(10)
		adminID := lfutils.GenerateRandomString(10)
		level := "Developer"
		expirationDate := time.Now().Add(time.Hour * 1)

		jwt, httpStatus, err := GenerateJwt(adminFullName, adminID, level, &expirationDate)
		assert.Equal(t, httpStatus, http.StatusOK)
		assert.Nil(t, err)

		headers := map[string]string{
			"Authorization": "Bearer " + jwt,
		}

		req := events.APIGatewayProxyRequest{
			HTTPMethod:     "GET",
			Headers:        headers,
			RequestContext: lfutils.GenerateAPIGatewayContextFromContext(),
		}

		jwtMiddlewareHandler := DecodeJWTMiddleware(getEmptyHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
		assert.Equal(t, res.StatusCode, http.StatusOK)
		// these headers aren't used in actual responses - they're here to
		//show the values are correctly injected from context during runtime
		assert.Equal(t, res.Headers["adminFullName"], adminFullName)
		assert.Equal(t, res.Headers["adminID"], adminID)
		assert.Equal(t, res.Headers["adminLevel"], level)

		expirationDateAsTime, err := time.Parse(time.RFC3339, res.Headers["expirationDate"])
		assert.Nil(t, err)

		assert.True(t, lfutils.AreSameSecond(&expirationDateAsTime, &expirationDate))
	})

	t.Run("OPTIONS req succeeds with invalid JWT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		adminFullName := lfutils.GenerateRandomString(10)
		adminID := lfutils.GenerateRandomString(10)
		level := "Developer"
		expirationDate := time.Now().Add(time.Hour * 1)

		jwt, httpStatus, err := GenerateJwt(adminFullName, adminID, level, &expirationDate)
		assert.Equal(t, httpStatus, http.StatusOK)
		assert.Nil(t, err)

		jwt = jwt + "hi" // create an invalid JWT

		headers := make(map[string]string)
		headers["Authorization"] = "Bearer " + jwt

		req := events.APIGatewayProxyRequest{
			HTTPMethod:     "OPTIONS",
			Headers:        headers,
			RequestContext: lfutils.GenerateAPIGatewayContextFromContext(),
		}

		jwtMiddlewareHandler := DecodeJWTMiddleware(getEmptyHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})

	t.Run("OPTIONS req succeeds with no Authorization header", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		req := events.APIGatewayProxyRequest{
			HTTPMethod:     "OPTIONS",
			Headers:        nil,
			RequestContext: lfutils.GenerateAPIGatewayContextFromContext(),
		}

		jwtMiddlewareHandler := DecodeJWTMiddleware(getEmptyHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})
}
