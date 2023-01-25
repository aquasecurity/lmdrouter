package lambda_jwt

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/jgroeneveld/trial/assert"
	"github.com/seantcanavan/lambda_jwt_router/response"
	"github.com/seantcanavan/lambda_jwt_router/router"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestAllowOptionsMW(t *testing.T) {
	t.Run("verify empty OPTIONS req succeeds", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodOptions,
			Headers:        nil,
			RequestContext: GenerateAPIGatewayProxyReq(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptionsMW(GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})
	t.Run("verify OPTIONS req succeeds with invalid JWT for AllowOptions", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		signedJWT, err := Sign(nil)
		assert.Nil(t, err)

		signedJWT = signedJWT + "hi" // create an invalid JWT

		req := events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodOptions,
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: GenerateAPIGatewayProxyReq(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptionsMW(GenerateEmptySuccessHandler())
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
			RequestContext: GenerateAPIGatewayProxyReq(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptionsMW(GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)
	})
}

func TestDecodeAndInjectExpandedClaims(t *testing.T) {
	t.Run("verify error is returned by DecodeAndInjectExpandedClaims when missing Authorization header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		jwtMiddlewareHandler := DecodeAndInjectExpandedClaims(GenerateEmptyErrorHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusBadRequest)

		var responseBody response.HTTPError
		err = router.UnmarshalResponse(res, &responseBody)
		assert.Nil(t, err)

		assert.Equal(t, responseBody.Status, res.StatusCode)
		assert.Equal(t, responseBody.Message, ErrNoAuthorizationHeader.Error())
	})
	t.Run("verify context is returned by DecodeAndInjectExpandedClaims with a signed JWT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		expandedClaims := generateExpandedMapClaims()

		signedJWT, err := Sign(expandedClaims)
		assert.Nil(t, err)

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: GenerateAPIGatewayProxyReq(),
		}

		jwtMiddlewareHandler := DecodeAndInjectExpandedClaims(GenerateSuccessHandlerAndMapExpandedContext())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		var returnedClaims ExpandedClaims
		err = router.UnmarshalResponse(res, &returnedClaims)
		assert.Nil(t, err)
		// this verifies that the context gets set in the middleware inject function since the
		// dummy handler passed to it as the 'next' call injects the values from its passed
		// context object into the response body. The function doesn't work this way in practice
		// however it does allow me to fully unit test it to make sure the context setting is working.
		// It's hacky and I'm not proud of it but I'm not sure how else to do it.
		assert.Equal(t, expandedClaims[AudienceKey], returnedClaims.Audience)
		assert.Equal(t, expandedClaims[ExpiresAtKey], returnedClaims.ExpiresAt)
		assert.Equal(t, expandedClaims[FirstNameKey], returnedClaims.FirstName)
		assert.Equal(t, expandedClaims[FullNameKey], returnedClaims.FullName)
		assert.Equal(t, expandedClaims[IDKey], returnedClaims.ID)
		assert.Equal(t, expandedClaims[IssuedAtKey], returnedClaims.IssuedAt)
		assert.Equal(t, expandedClaims[IssuerKey], returnedClaims.Issuer)
		assert.Equal(t, expandedClaims[LevelKey], returnedClaims.Level)
		assert.Equal(t, expandedClaims[NotBeforeKey], returnedClaims.NotBefore)
		assert.Equal(t, expandedClaims[SubjectKey], returnedClaims.Subject)
		assert.Equal(t, expandedClaims[UserTypeKey], returnedClaims.UserType)
	})
}

func TestDecodeAndInjectStandardClaims(t *testing.T) {
	t.Run("verify error is returned by DecodeAndInjectStandardClaims when missing Authorization header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		jwtMiddlewareHandler := DecodeAndInjectStandardClaims(GenerateEmptyErrorHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusBadRequest)

		var responseBody response.HTTPError
		err = router.UnmarshalResponse(res, &responseBody)
		assert.Nil(t, err)

		assert.Equal(t, responseBody.Status, res.StatusCode)
		assert.Equal(t, responseBody.Message, ErrNoAuthorizationHeader.Error())
	})
	t.Run("verify context is returned by DecodeAndInjectStandardClaims with a signed JWT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		standardClaims := generateStandardMapClaims()

		signedJWT, err := Sign(standardClaims)
		assert.Nil(t, err)

		req := events.APIGatewayProxyRequest{
			HTTPMethod: "GET",
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: GenerateAPIGatewayProxyReq(),
		}

		jwtMiddlewareHandler := DecodeAndInjectStandardClaims(GenerateSuccessHandlerAndMapStandardContext())
		res, err := jwtMiddlewareHandler(ctx, req)
		assert.Nil(t, err)
		assert.Equal(t, res.StatusCode, http.StatusOK)

		var returnedClaims jwt.StandardClaims
		err = router.UnmarshalResponse(res, &returnedClaims)
		assert.Nil(t, err)
		// this verifies that the context gets set in the middleware inject function since the
		// dummy handler passed to it as the 'next' call injects the values from its passed
		// context object into the response body. The function doesn't work this way in practice
		// however it does allow me to fully unit test it to make sure the context setting is working.
		// It's hacky and I'm not proud of it but I'm not sure how else to do it.
		assert.Equal(t, returnedClaims.Audience, standardClaims[AudienceKey])
		assert.Equal(t, returnedClaims.ExpiresAt, standardClaims[ExpiresAtKey])
		assert.Equal(t, returnedClaims.Id, standardClaims[IDKey])
		assert.Equal(t, returnedClaims.IssuedAt, standardClaims[IssuedAtKey])
		assert.Equal(t, returnedClaims.Issuer, standardClaims[IssuerKey])
		assert.Equal(t, returnedClaims.NotBefore, standardClaims[NotBeforeKey])
		assert.Equal(t, returnedClaims.Subject, standardClaims[SubjectKey])
	})
}

func TestExtractJWT(t *testing.T) {
	standardClaims := generateStandardMapClaims()
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
func GenerateSuccessHandlerAndMapExpandedContext() router.Handler {
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
func GenerateSuccessHandlerAndMapStandardContext() router.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return response.Custom(http.StatusOK, nil, jwt.StandardClaims{
			Audience:  ctx.Value(AudienceKey).(string),
			ExpiresAt: ctx.Value(ExpiresAtKey).(int64),
			Id:        ctx.Value(IDKey).(string),
			IssuedAt:  ctx.Value(IssuedAtKey).(int64),
			Issuer:    ctx.Value(IssuerKey).(string),
			NotBefore: ctx.Value(NotBeforeKey).(int64),
			Subject:   ctx.Value(SubjectKey).(string),
		})
	}
}

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

func GenerateEmptySuccessHandler() router.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return response.Empty()
	}
}

func GenerateEmptyErrorHandler() router.Handler {
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
