package middleware

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/seantcanavan/lmdrouter"
	"github.com/seantcanavan/lmdrouter/jwt_auth"
	"github.com/seantcanavan/lmdrouter/response"
	"log"
	"net/http"
)

// LogRequest is a standard middleware function that will log every incoming
// events.APIGatewayProxyRequest request and the pertinent information in it.
func LogRequest(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		format := "[%s] [%s %s] [%d]%s"
		level := "INF"
		var code int
		var extra string

		res, err = next(ctx, req)
		if err != nil {
			level = "ERR"
			code = http.StatusInternalServerError
			extra = " " + err.Error()
		} else {
			code = res.StatusCode
			if code >= 400 {
				level = "ERR"
			}
		}

		log.Printf(format, level, req.HTTPMethod, req.Path, code, extra)

		return res, err
	}
}

// AllowOptions is a helper middleware function that will immediately return
// a successful request if the method is OPTIONS. This makes sure that
// HTTP OPTIONS calls for CORS functionality are supported.
func AllowOptions(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		if req.HTTPMethod == "OPTIONS" { // immediately return success for options calls for CORS reqs
			return response.Empty()
		}

		return next(ctx, req)
	}
}

// DecodeAndInjectStandardClaims attempts to parse a Json Web Token from the req's "Authorization"
// header. If the Authorization header is missing, or does not contain a valid Json Web Token
// (JWT) then an error message and appropriate HTTP status code will be returned. If the JWT
// is correctly set and contains a StandardClaim then the values from that standard claim
// will be added to the context object for others to use during their processing.
func DecodeAndInjectStandardClaims(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		mapClaims, httpStatus, err := jwt_auth.ExtractJWT(req.Headers)
		if err != nil {
			return response.ErrorAndStatus(httpStatus, err)
		}

		var standardClaims jwt.StandardClaims
		err = jwt_auth.ExtractStandardClaims(mapClaims, standardClaims)
		if err != nil {
			return response.ErrorAndStatus(http.StatusInternalServerError, err)
		}

		ctx = context.WithValue(ctx, jwt_auth.AudienceKey, standardClaims.Audience)
		ctx = context.WithValue(ctx, jwt_auth.ExpiresAtKey, standardClaims.ExpiresAt)
		ctx = context.WithValue(ctx, jwt_auth.IDKey, standardClaims.Id)
		ctx = context.WithValue(ctx, jwt_auth.IssuedAtKey, standardClaims.IssuedAt)
		ctx = context.WithValue(ctx, jwt_auth.IssuerKey, standardClaims.Issuer)
		ctx = context.WithValue(ctx, jwt_auth.NotBeforeKey, standardClaims.NotBefore)
		ctx = context.WithValue(ctx, jwt_auth.SubjectKey, standardClaims.Subject)

		res, err = next(ctx, req)
		return res, err
	}
}

// DecodeAndInjectExpandedClaims attempts to parse a Json Web Token from the req's "Authorization"
// header. If the Authorization header is missing, or does not contain a valid Json Web Token
// (JWT) then an error message and appropriate HTTP status code will be returned. If the JWT
// is correctly set and contains an instance of jwt_auth.ExpandedClaims then the values from
// that standard claim will be added to the context object for others to use during their processing.
func DecodeAndInjectExpandedClaims(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		mapClaims, httpStatus, err := jwt_auth.ExtractJWT(req.Headers)
		if err != nil {
			return response.ErrorAndStatus(httpStatus, err)
		}

		var extendedClaims jwt_auth.ExpandedClaims
		err = jwt_auth.ExtractCustomClaims(mapClaims, &extendedClaims)
		if err != nil {
			return response.ErrorAndStatus(http.StatusInternalServerError, err)
		}

		ctx = context.WithValue(ctx, jwt_auth.AudienceKey, extendedClaims.Audience)
		ctx = context.WithValue(ctx, jwt_auth.ExpiresAtKey, extendedClaims.ExpiresAt)
		ctx = context.WithValue(ctx, jwt_auth.FirstNameKey, extendedClaims.FirstName)
		ctx = context.WithValue(ctx, jwt_auth.FullNameKey, extendedClaims.FullName)
		ctx = context.WithValue(ctx, jwt_auth.IDKey, extendedClaims.ID)
		ctx = context.WithValue(ctx, jwt_auth.IssuedAtKey, extendedClaims.IssuedAt)
		ctx = context.WithValue(ctx, jwt_auth.IssuerKey, extendedClaims.Issuer)
		ctx = context.WithValue(ctx, jwt_auth.LevelKey, extendedClaims.Level)
		ctx = context.WithValue(ctx, jwt_auth.NotBeforeKey, extendedClaims.NotBefore)
		ctx = context.WithValue(ctx, jwt_auth.SubjectKey, extendedClaims.Subject)
		ctx = context.WithValue(ctx, jwt_auth.UserTypeKey, extendedClaims.UserType)

		res, err = next(ctx, req)
		return res, err
	}
}
