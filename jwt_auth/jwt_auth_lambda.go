package jwt_auth

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/seantcanavan/lmdrouter"
	"github.com/seantcanavan/lmdrouter/response"
	"net/http"
	"strings"
)

var ErrNoAuthorizationHeader = errors.New("no Authorization header value set")
var ErrNoBearerPrefix = errors.New("missing 'Bearer ' prefix for Authorization header value")
var ErrVerifyJWT = errors.New("unable to verify JWT to retrieve claims. try logging in again to ensure it is not expired")

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
		mapClaims, httpStatus, err := ExtractJWT(req.Headers)
		if err != nil {
			return response.ErrorAndStatus(httpStatus, err)
		}

		var standardClaims jwt.StandardClaims
		err = ExtractStandardClaims(mapClaims, standardClaims)
		if err != nil {
			return response.ErrorAndStatus(http.StatusInternalServerError, err)
		}

		ctx = context.WithValue(ctx, AudienceKey, standardClaims.Audience)
		ctx = context.WithValue(ctx, ExpiresAtKey, standardClaims.ExpiresAt)
		ctx = context.WithValue(ctx, IDKey, standardClaims.Id)
		ctx = context.WithValue(ctx, IssuedAtKey, standardClaims.IssuedAt)
		ctx = context.WithValue(ctx, IssuerKey, standardClaims.Issuer)
		ctx = context.WithValue(ctx, NotBeforeKey, standardClaims.NotBefore)
		ctx = context.WithValue(ctx, SubjectKey, standardClaims.Subject)

		res, err = next(ctx, req)
		return res, err
	}
}

// DecodeAndInjectExpandedClaims attempts to parse a Json Web Token from the req's "Authorization"
// header. If the Authorization header is missing, or does not contain a valid Json Web Token
// (JWT) then an error message and appropriate HTTP status code will be returned. If the JWT
// is correctly set and contains an instance of ExpandedClaims then the values from
// that standard claim will be added to the context object for others to use during their processing.
func DecodeAndInjectExpandedClaims(next lmdrouter.Handler) lmdrouter.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		mapClaims, httpStatus, err := ExtractJWT(req.Headers)
		if err != nil {
			return response.ErrorAndStatus(httpStatus, err)
		}

		var extendedClaims ExpandedClaims
		err = ExtractCustomClaims(mapClaims, &extendedClaims)
		if err != nil {
			return response.ErrorAndStatus(http.StatusInternalServerError, err)
		}

		ctx = context.WithValue(ctx, AudienceKey, extendedClaims.Audience)
		ctx = context.WithValue(ctx, ExpiresAtKey, extendedClaims.ExpiresAt)
		ctx = context.WithValue(ctx, FirstNameKey, extendedClaims.FirstName)
		ctx = context.WithValue(ctx, FullNameKey, extendedClaims.FullName)
		ctx = context.WithValue(ctx, IDKey, extendedClaims.ID)
		ctx = context.WithValue(ctx, IssuedAtKey, extendedClaims.IssuedAt)
		ctx = context.WithValue(ctx, IssuerKey, extendedClaims.Issuer)
		ctx = context.WithValue(ctx, LevelKey, extendedClaims.Level)
		ctx = context.WithValue(ctx, NotBeforeKey, extendedClaims.NotBefore)
		ctx = context.WithValue(ctx, SubjectKey, extendedClaims.Subject)
		ctx = context.WithValue(ctx, UserTypeKey, extendedClaims.UserType)

		res, err = next(ctx, req)
		return res, err
	}
}

// ExtractJWT will attempt to extract the JWT value and retrieve the map claims from an
// events.APIGatewayProxyRequest object. If there is an error that will be returned
// along with an appropriate HTTP status code as an integer. If everything goes right
// then error will be nil and the int will be http.StatusOK
func ExtractJWT(headers map[string]string) (jwt.MapClaims, int, error) {
	authorizationHeader := headers["Authorization"]
	if authorizationHeader == "" {
		return nil, http.StatusBadRequest, ErrNoAuthorizationHeader
	}

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, http.StatusBadRequest, ErrNoBearerPrefix
	}

	userJwt := strings.TrimPrefix(authorizationHeader, "Bearer ")

	mapClaims, err := VerifyJWT(userJwt)
	if err != nil {
		return nil, http.StatusUnauthorized, ErrVerifyJWT
	}

	return mapClaims, http.StatusOK, nil
}
