package lambda_jwt

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/golang-jwt/jwt"
	"github.com/seantcanavan/lambda_jwt_router/response"
	"github.com/seantcanavan/lambda_jwt_router/router"
	"log"
	"net/http"
	"strings"
)

var ErrNoAuthorizationHeader = errors.New("no Authorization header value set")
var ErrNoBearerPrefix = errors.New("missing 'Bearer ' prefix for Authorization header value")
var ErrVerifyJWT = errors.New("unable to verify JWT to retrieve claims. try logging in again to ensure it is not expired")

// AllowOptionsMW is a helper middleware function that will immediately return
// a successful request if the method is OPTIONS. This makes sure that
// HTTP OPTIONS calls for CORS functionality are supported.
func AllowOptionsMW(next router.Handler) router.Handler {
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

// DecodeAndInjectStandardClaims attempts to parse a Json Web Token from the request's "Authorization"
// header. If the Authorization header is missing, or does not contain a valid Json Web Token
// (JWT) then an error message and appropriate HTTP status code will be returned. If the JWT
// is correctly set and contains a StandardClaim then the values from that standard claim
// will be added to the context object for others to use during their processing.
func DecodeAndInjectStandardClaims(next router.Handler) router.Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		mapClaims, httpStatus, err := ExtractJWT(req.Headers)
		if err != nil {
			return response.ErrorAndStatus(httpStatus, err)
		}

		var standardClaims jwt.StandardClaims
		err = ExtractStandardClaims(mapClaims, &standardClaims)
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

// DecodeAndInjectExpandedClaims attempts to parse a Json Web Token from the request's "Authorization"
// header. If the Authorization header is missing, or does not contain a valid Json Web Token
// (JWT) then an error message and appropriate HTTP status code will be returned. If the JWT
// is correctly set and contains an instance of ExpandedClaims then the values from
// that standard claim will be added to the context object for others to use during their processing.
func DecodeAndInjectExpandedClaims(next router.Handler) router.Handler {
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

// LogRequestMW is a standard middleware function that will log every incoming
// events.APIGatewayProxyRequest request and the pertinent information in it.
//goland:noinspection GoUnusedExportedFunction
func LogRequestMW(next router.Handler) router.Handler {
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
