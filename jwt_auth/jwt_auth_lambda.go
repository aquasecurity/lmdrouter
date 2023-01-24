package jwt_auth

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"net/http"
	"strings"
)

var ErrNoAuthorizationHeader = errors.New("no Authorization header value set")
var ErrNoBearerPrefix = errors.New("missing 'Bearer ' prefix for Authorization header value")
var ErrVerifyJWT = errors.New("unable to verify JWT to retrieve claims. try logging in again to ensure it is not expired")

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
