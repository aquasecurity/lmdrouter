package jwt_auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/jgroeneveld/trial/assert"
	"github.com/seantcanavan/lmdrouter"
	"github.com/seantcanavan/lmdrouter/response"
	"github.com/seantcanavan/lmdrouter/utils"
	"net/http"
	"testing"
	"time"
)

func TestDecodeAndInjectExpandedClaims(t *testing.T) {
	t.Run("verify error is returned by DecodeAndInjectExpandedClaims when missing Authorization header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		jwtMiddlewareHandler := DecodeAndInjectExpandedClaims(utils.GenerateEmptyErrorHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
		assert.Equal(t, res.StatusCode, http.StatusBadRequest)

		var responseBody response.HTTPError
		err = json.Unmarshal([]byte(res.Body), &responseBody)
		assert.Nil(t, err)

		assert.Equal(t, responseBody.Status, res.StatusCode)
		assert.Equal(t, responseBody.Message, ErrNoAuthorizationHeader.Error())
	})
	t.Run("verify call to DecodeAndInjectExpandedClaims with a signed JWT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		signedJWT, err := Sign(jwt.MapClaims{"hello": "there"})
		assert.Nil(t, err)

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: utils.GenerateAPIGatewayProxyReq(),
		}

		jwtMiddlewareHandler := DecodeAndInjectExpandedClaims(utils.GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(ctx, req)

		fmt.Println(fmt.Sprintf("ctx is [%+v]", ctx))
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})
}

func TestDecodeAndInjectStandardClaims(t *testing.T) {
	t.Run("verify error is returned by DecodeAndInjectStandardClaims when missing Authorization header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		jwtMiddlewareHandler := DecodeAndInjectExpandedClaims(utils.GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		assert.Nil(t, err) // because the req didn't crash the stack - error is nil
		assert.Equal(t, res.StatusCode, http.StatusBadRequest)

		var responseBody response.HTTPError
		err = json.Unmarshal([]byte(res.Body), &responseBody)
		assert.Nil(t, err)

		assert.Equal(t, responseBody.Status, res.StatusCode)
		assert.Equal(t, responseBody.Message, ErrNoAuthorizationHeader.Error())
	})
	t.Run("verify context is returned by DecodeAndInjectStandardClaims with a signed JWT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		standardClaims := GenerateStandardMapClaims()

		signedJWT, err := Sign(standardClaims)
		assert.Nil(t, err)

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: utils.GenerateAPIGatewayProxyReq(),
		}

		jwtMiddlewareHandler := DecodeAndInjectExpandedClaims(GenerateSuccessHandlerAndMapStandardContext())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		var returnedClaims jwt.StandardClaims
		err = lmdrouter.UnmarshalResponse(res, &returnedClaims)
		assert.Nil(t, err)
		assert.Equal(t, standardClaims, returnedClaims)
	})
}

func TestExtractJWT(t *testing.T) {
	standardClaims := GenerateStandardMapClaims()
	signedJWT, err := Sign(standardClaims)
	assert.Nil(t, err)

	t.Run("verify ExtractJWT returns err for empty Authorization header", func(t *testing.T) {
		headers := map[string]string{"Authorization": ""}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for Authorization header misspelled - all caps", func(t *testing.T) {
		headers := map[string]string{"AUTHORIZATION": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for Authorization header misspelled - lowercase", func(t *testing.T) {
		headers := map[string]string{"authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for bearer prefix not used", func(t *testing.T) {
		headers := map[string]string{"Authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for bearer not camel cased", func(t *testing.T) {
		headers := map[string]string{"Authorization": "bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for BEARER all caps", func(t *testing.T) {
		headers := map[string]string{"Authorization": "BEARER " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for Bearer does not end with space", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer" + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusBadRequest, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns claims correctly with valid input", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 7)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.Nil(t, extractErr)
		assert.Nil(t, extractErr)

		assert.Equal(t, mapClaims[AudienceKey], mapClaims[AudienceKey])
		assert.Equal(t, mapClaims[ExpiresAtKey], mapClaims[ExpiresAtKey])
		assert.Equal(t, mapClaims[IDKey], mapClaims[IDKey])
		assert.Equal(t, mapClaims[IssuedAtKey], mapClaims[IssuedAtKey])
		assert.Equal(t, mapClaims[IssuerKey], mapClaims[IssuerKey])
		assert.Equal(t, mapClaims[NotBeforeKey], mapClaims[NotBeforeKey])
		assert.Equal(t, mapClaims[SubjectKey], mapClaims[SubjectKey])
	})
}

// GenerateSuccessHandlerAndMapExpandedContext returns a middleware handler
// that takes the values inserted into the context object by DecodeAndInjectExpandedClaims
// and returns them as an object from the request so that unit tests can analyze the values
// and make sure they have done the full trip from JWT -> CTX -> unit test
func GenerateSuccessHandlerAndMapExpandedContext() lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return response.Custom(http.StatusOK, nil, ExpandedClaims{
			Audience:  ctx.Value(AudienceKey).(string),
			ExpiresAt: ctx.Value(ExpiresAtKey).(int64),
			FirstName: ctx.Value(FirstNameKey).(string),
			FullName:  ctx.Value(FullNameKey).(string),
			ID:        ctx.Value(IDKey).(string),
			IssuedAt:  ctx.Value(IssuedAtKey).(int64),
			Issuer:    ctx.Value(IssuerKey).(string),
			Level:     ctx.Value(LevelKey).(string),
			NotBefore: ctx.Value(NotBeforeKey).(int64),
			Subject:   ctx.Value(SubjectKey).(string),
			UserType:  ctx.Value(UserTypeKey).(string),
		})
	}
}

// GenerateSuccessHandlerAndMapStandardContext returns a middleware handler
// that takes the values inserted into the context object by DecodeAndInjectStandardClaims
// and returns them as an object from the request so that unit tests can analyze the values
// and make sure they have done the full trip from JWT -> CTX -> unit test
func GenerateSuccessHandlerAndMapStandardContext() lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return response.Custom(http.StatusOK, nil, jwt.StandardClaims{
			Audience:  ctx.Value(AudienceKey).(string),
			ExpiresAt: ctx.Value(ExpiresAtKey).(int64),
			IssuedAt:  ctx.Value(IssuedAtKey).(int64),
			Issuer:    ctx.Value(IssuerKey).(string),
			NotBefore: ctx.Value(NotBeforeKey).(int64),
			Subject:   ctx.Value(SubjectKey).(string),
		})
	}
}
