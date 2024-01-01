package lambda_jwt

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/seantcanavan/lambda_jwt_router/lambda_router"
	"github.com/seantcanavan/lambda_jwt_router/lambda_util"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestAllowOptionsMW(t *testing.T) {
	t.Run("verify empty OPTIONS req succeeds", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodOptions,
			Headers:        nil,
			RequestContext: lambda_util.GenerateRandomAPIGatewayContext(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptionsMW(GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusOK)
	})
	t.Run("verify OPTIONS req succeeds with invalid JWT for AllowOptions", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		signedJWT, err := Sign(nil)
		require.Nil(t, err)

		signedJWT = signedJWT + "hi" // create an invalid JWT

		req := events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodOptions,
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: lambda_util.GenerateRandomAPIGatewayContext(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptionsMW(GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusOK)
	})
	t.Run("verify OPTIONS req succeeds with no Authorization header for AllowOptions", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		req := events.APIGatewayProxyRequest{
			HTTPMethod:     http.MethodOptions,
			Headers:        nil,
			RequestContext: lambda_util.GenerateRandomAPIGatewayContext(),
		}

		// we pass along an error handler but expect http.StatusOK because the AllowOptions handler should execute first
		jwtMiddlewareHandler := AllowOptionsMW(GenerateEmptySuccessHandler())
		res, err := jwtMiddlewareHandler(ctx, req)
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusOK)
	})
}

func TestDecodeAndInjectExpandedClaims(t *testing.T) {
	t.Run("verify error is returned by DecodeExpanded when missing Authorization header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		jwtMiddlewareHandler := DecodeExpanded(GenerateEmptyErrorHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusBadRequest)

		var responseBody lambda_router.HTTPError
		err = lambda_router.UnmarshalRes(res, &responseBody)
		require.Nil(t, err)

		require.Equal(t, responseBody.Status, res.StatusCode)
		require.Equal(t, responseBody.Message, ErrNoAuthorizationHeader.Error())
	})
	t.Run("verify context is returned by DecodeExpanded with a signed JWT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		expandedClaims := generateExpandedMapClaims()

		signedJWT, err := Sign(expandedClaims)
		require.Nil(t, err)

		req := events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodGet,
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: lambda_util.GenerateRandomAPIGatewayContext(),
		}

		jwtMiddlewareHandler := DecodeExpanded(generateSuccessHandlerAndMapExpandedContext())
		res, err := jwtMiddlewareHandler(ctx, req)
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusOK)

		var returnedClaims ExpandedClaims
		err = lambda_router.UnmarshalRes(res, &returnedClaims)
		require.Nil(t, err)
		// this verifies that the context gets set in the middleware inject function since the
		// dummy handler passed to it as the 'next' call injects the values from its passed
		// context object into the response body. The function doesn't work this way in practice
		// however it does allow me to fully unit test it to make sure the context setting is working.
		// It's hacky and I'm not proud of it but I'm not sure how else to do it.
		require.Equal(t, expandedClaims[AudienceKey], returnedClaims.Audience)
		require.Equal(t, expandedClaims[EmailKey], returnedClaims.Email)
		require.Equal(t, expandedClaims[ExpiresAtKey], returnedClaims.ExpiresAt)
		require.Equal(t, expandedClaims[FirstNameKey], returnedClaims.FirstName)
		require.Equal(t, expandedClaims[FullNameKey], returnedClaims.FullName)
		require.Equal(t, expandedClaims[IDKey], returnedClaims.ID)
		require.Equal(t, expandedClaims[IssuedAtKey], returnedClaims.IssuedAt)
		require.Equal(t, expandedClaims[IssuerKey], returnedClaims.Issuer)
		require.Equal(t, expandedClaims[LevelKey], returnedClaims.Level)
		require.Equal(t, expandedClaims[NotBeforeKey], returnedClaims.NotBefore)
		require.Equal(t, expandedClaims[SubjectKey], returnedClaims.Subject)
		require.Equal(t, expandedClaims[UserTypeKey], returnedClaims.UserType)
	})
}

func TestDecodeAndInjectStandardClaims(t *testing.T) {
	t.Run("verify error is returned by DecodeStandard when missing Authorization header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		jwtMiddlewareHandler := DecodeStandard(GenerateEmptyErrorHandler())
		res, err := jwtMiddlewareHandler(nil, req)
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusBadRequest)

		var responseBody lambda_router.HTTPError
		err = lambda_router.UnmarshalRes(res, &responseBody)
		require.Nil(t, err)

		require.Equal(t, responseBody.Status, res.StatusCode)
		require.Equal(t, responseBody.Message, ErrNoAuthorizationHeader.Error())
	})
	t.Run("verify context is returned by DecodeStandard with a signed JWT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()

		standardClaims := generateStandardMapClaims()

		signedJWT, err := Sign(standardClaims)
		require.Nil(t, err)

		req := events.APIGatewayProxyRequest{
			HTTPMethod: http.MethodGet,
			Headers: map[string]string{
				"Authorization": "Bearer " + signedJWT,
			},
			RequestContext: lambda_util.GenerateRandomAPIGatewayContext(),
		}

		jwtMiddlewareHandler := DecodeStandard(generateSuccessHandlerAndMapStandardContext())
		res, err := jwtMiddlewareHandler(ctx, req)
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusOK)

		var returnedClaims jwt.StandardClaims
		err = lambda_router.UnmarshalRes(res, &returnedClaims)
		require.Nil(t, err)
		// this verifies that the context gets set in the middleware inject function since the
		// dummy handler passed to it as the 'next' call injects the values from its passed
		// context object into the response body. The function doesn't work this way in practice
		// however it does allow me to fully unit test it to make sure the context setting is working.
		// It's hacky and I'm not proud of it but I'm not sure how else to do it.
		require.Equal(t, returnedClaims.Audience, standardClaims[AudienceKey])
		require.Equal(t, returnedClaims.ExpiresAt, standardClaims[ExpiresAtKey])
		require.Equal(t, returnedClaims.Id, standardClaims[IDKey])
		require.Equal(t, returnedClaims.IssuedAt, standardClaims[IssuedAtKey])
		require.Equal(t, returnedClaims.Issuer, standardClaims[IssuerKey])
		require.Equal(t, returnedClaims.NotBefore, standardClaims[NotBeforeKey])
		require.Equal(t, returnedClaims.Subject, standardClaims[SubjectKey])
	})
}

func TestExtractJWT(t *testing.T) {
	standardClaims := generateStandardMapClaims()
	signedJWT, err := Sign(standardClaims)
	require.Nil(t, err)

	t.Run("verify ExtractJWT returns err for empty Authorization header", func(t *testing.T) {
		headers := map[string]string{"Authorization": ""}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for Authorization header misspelled - all caps", func(t *testing.T) {
		headers := map[string]string{"AUTHORIZATION": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for Authorization header misspelled - lowercase", func(t *testing.T) {
		headers := map[string]string{"authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for bearer prefix not used", func(t *testing.T) {
		headers := map[string]string{"Authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for bearer not camel cased", func(t *testing.T) {
		headers := map[string]string{"Authorization": "bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for BEARER all caps", func(t *testing.T) {
		headers := map[string]string{"Authorization": "BEARER " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for Bearer does not end with space", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer" + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns claims correctly with valid input", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 7)
		require.Equal(t, http.StatusOK, httpStatus)
		require.Nil(t, extractErr)
		require.Nil(t, extractErr)

		require.Equal(t, mapClaims[AudienceKey], mapClaims[AudienceKey])
		require.Equal(t, mapClaims[ExpiresAtKey], mapClaims[ExpiresAtKey])
		require.Equal(t, mapClaims[IDKey], mapClaims[IDKey])
		require.Equal(t, mapClaims[IssuedAtKey], mapClaims[IssuedAtKey])
		require.Equal(t, mapClaims[IssuerKey], mapClaims[IssuerKey])
		require.Equal(t, mapClaims[NotBeforeKey], mapClaims[NotBeforeKey])
		require.Equal(t, mapClaims[SubjectKey], mapClaims[SubjectKey])
	})
}

func TestGenerateEmptyErrorHandler(t *testing.T) {
	t.Run("verify empty error handler returns error", func(t *testing.T) {
		errHandler := GenerateEmptyErrorHandler()
		res, err := errHandler(nil, lambda_util.GenerateRandomAPIGatewayProxyRequest())
		require.Nil(t, err) // err handler embeds the error in the response, not the golang stack
		require.Equal(t, res.StatusCode, http.StatusInternalServerError)
		var httpError lambda_router.HTTPError
		err = lambda_router.UnmarshalRes(res, &httpError)
		require.Nil(t, err)
		require.Equal(t, httpError.Status, http.StatusInternalServerError)
		require.Equal(t, httpError.Message, "this error is simulated")
	})
}

func TestGenerateEmptySuccessHandler(t *testing.T) {
	t.Run("verify empty success handler returns success", func(t *testing.T) {
		successHandler := GenerateEmptySuccessHandler()
		res, err := successHandler(nil, lambda_util.GenerateRandomAPIGatewayProxyRequest())
		require.Nil(t, err)
		require.Equal(t, res.StatusCode, http.StatusOK)
		require.Equal(t, res.Body, "{}") // empty struct response
	})
}

// generateSuccessHandlerAndMapExpandedContext returns a middleware handler
// that takes the values inserted into the context object by DecodeExpanded
// and returns them as an object from the request so that unit tests can analyze the values
// and make sure they have done the full trip from JWT -> CTX -> unit test
func generateSuccessHandlerAndMapExpandedContext() lambda_router.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return lambda_router.CustomRes(http.StatusOK, nil, ExpandedClaims{
			Audience:  ctx.Value(AudienceKey).(string),
			Email:     ctx.Value(EmailKey).(string),
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

// generateSuccessHandlerAndMapStandardContext returns a middleware handler
// that takes the values inserted into the context object by DecodeStandard
// and returns them as an object from the request so that unit tests can analyze the values
// and make sure they have done the full trip from JWT -> CTX -> unit test
func generateSuccessHandlerAndMapStandardContext() lambda_router.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		events.APIGatewayProxyResponse,
		error) {
		return lambda_router.CustomRes(http.StatusOK, nil, jwt.StandardClaims{
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
