// Package ljwt appends critical libraries necessary for using JWTs (Json Web Tokens)
// within AWS Lambda through API Gateway proxy requests / integration. It adds multiple
// middleware functions for checking and validating permissions based on user type and
// has multiple examples of appending information from the caller's JWT directly into
// the golang context object so other handler functions can utilize the information.
// If you wish to use the standard 7 JWT values as defined by Auth0 at
// https://auth0.com/docs/secure/tokens/json-web-tokens/json-web-token-claims
// then you want to use the jwt.StandardClaims object. If you wish to use an
// expanded claim set with a few additional helpful values like email and usertype
// then check out the ExpandedClaims object. If you wish to provide your own
// totally custom claim values and object then check out ExtractCustom.
package ljwt

import (
	"encoding/hex"
	"encoding/json"
	"github.com/golang-jwt/jwt"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/seantcanavan/lambda_jwt_router/lcom"
	"log"
	"net/http"
	"os"
	"strings"
)

type ExpandedClaims struct {
	Audience  string `json:"aud"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"exp"`
	FirstName string `json:"firstName"`
	FullName  string `json:"fullName"`
	ID        string `json:"jti"`
	IssuedAt  int64  `json:"iat"`
	Issuer    string `json:"iss"`
	Level     string `json:"level"`
	NotBefore int64  `json:"nbf"`
	Subject   string `json:"sub"`
	UserType  string `json:"userType"`
}

// ExtendExpanded returns an instance of jwt.MapClaims which you can freely extend
// with your own custom fields. It uses ExpandedClaims as the base struct to start with
// and returns a jwt.MapClaims which is just a wrapper for a map so you can add as many
// custom fields as you would like while still getting the 7 standard JWT fields and the
// 4 non-standard fields defined in this library.
func ExtendExpanded(claims ExpandedClaims) jwt.MapClaims {
	return jwt.MapClaims{
		lcom.JWTClaimAudienceKey:  claims.Audience,
		lcom.JWTClaimEmailKey:     claims.Email,
		lcom.JWTClaimExpiresAtKey: claims.ExpiresAt,
		lcom.JWTClaimFirstNameKey: claims.FirstName,
		lcom.JWTClaimFullNameKey:  claims.FullName,
		lcom.JWTClaimIDKey:        claims.ID,
		lcom.JWTClaimIssuedAtKey:  claims.IssuedAt,
		lcom.JWTClaimIssuerKey:    claims.Issuer,
		lcom.JWTClaimLevelKey:     claims.Level,
		lcom.JWTClaimNotBeforeKey: claims.NotBefore,
		lcom.JWTClaimSubjectKey:   claims.Subject,
		lcom.JWTClaimUserTypeKey:  claims.UserType,
	}
}

// ExtendStandard returns an instance of jwt.MapClaims which you can freely extend
// with your own custom fields. It uses jwt.StandardClaims as the base struct to start with
// and returns a jwt.MapClaims which is just a wrapper for a map so you can add as many
// custom fields as you would like while still getting the 7 standard JWT fields.
func ExtendStandard(claims jwt.StandardClaims) jwt.MapClaims {
	return jwt.MapClaims{
		lcom.JWTClaimAudienceKey:  claims.Audience,
		lcom.JWTClaimExpiresAtKey: claims.ExpiresAt,
		lcom.JWTClaimIDKey:        claims.Id,
		lcom.JWTClaimIssuedAtKey:  claims.IssuedAt,
		lcom.JWTClaimIssuerKey:    claims.Issuer,
		lcom.JWTClaimNotBeforeKey: claims.NotBefore,
		lcom.JWTClaimSubjectKey:   claims.Subject,
	}
}

// ExtractCustom takes in a generic claims map that can have any values
// set and attempts to pull out whatever custom struct you should have
// previously used to create the claims originally. An error will be
// returned if the generic map that stores the claims can't be converted
// to the struct of your choice through JSON marshalling.
func ExtractCustom(mapClaims jwt.MapClaims, val any) error {
	jsonBytes, err := json.Marshal(mapClaims)
	if err != nil {
		return util.WrapErrors(err, lcom.ErrMarshalMapClaims)
	}

	err = json.Unmarshal(jsonBytes, &val)
	if err != nil {
		return util.WrapErrors(err, lcom.ErrBadClaimsObject)
	}

	return nil
}

// ExtractStandard accepts a generic claims map that can have any values set and
// attempts to pull out a standard jwt.StandardClaims object from the claims map.
// The input claims should have been generated originally by a jwt.StandardClaims
// instance so they can be cleanly extracted back into an instance of jwt.StandardClaims.
func ExtractStandard(mapClaims jwt.MapClaims, standardClaims *jwt.StandardClaims) error {
	jsonBytes, err := json.Marshal(mapClaims)
	if err != nil {
		return util.WrapErrors(err, lcom.ErrMarshalMapClaims)
	}

	err = json.Unmarshal(jsonBytes, standardClaims)
	if err != nil {
		return util.WrapErrors(err, lcom.ErrBadClaimsObject)
	}

	return nil
}

// Sign accepts a final set of claims, either jwt.StandardClaims, ExpandedClaims,
// or something entirely custom that you have created yourself. It will sign the
// claims using the HMAC value loaded from environment variables and return the
// signed JWT if no error, otherwise the empty string and an error. To convert
// a GoLang struct to a claims object use ExtendStandard or ExtendExpanded
// to get started.
func Sign(mapClaims jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, mapClaims)

	// Sign and get the complete encoded token as a string using the secret
	encodedToken, err := token.SignedString(getBinarySecret())
	if err != nil {
		return "", util.WrapErrors(err, lcom.ErrUnableToSignToken)
	}

	return encodedToken, nil
}

// VerifyJWT accepts the user JWT from the Authorization header
// and returns the MapClaims or nil and an error set.
func VerifyJWT(userJWT string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(userJWT, keyFunc)
	if err != nil {
		return nil, util.WrapErrors(err, lcom.ErrInvalidJWT)
	}

	if !token.Valid {
		return nil, lcom.ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, lcom.ErrInvalidTokenClaims
}

func getBinarySecret() []byte {
	secret := os.Getenv(lcom.HMACSecretEnvKey)
	if secret == "" {
		log.Fatalf("cannot encode / decode with an empty secret")
	}

	data, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatalf("cannot decode the secret")
	}

	return data
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, lcom.ErrUnsupportedSigningMethod
	}

	return getBinarySecret(), nil
}

// ExtractJWT will attempt to extract the JWT value and retrieve the map claims from an
// events.APIGatewayProxyRequest object. If there is an error that will be returned
// along with an appropriate HTTP status code as an integer. If everything goes right
// then error will be nil and the int will be http.StatusOK
func ExtractJWT(headers map[string]string) (jwt.MapClaims, int, error) {
	authorizationHeader := headers["Authorization"]
	if authorizationHeader == "" {
		return nil, http.StatusBadRequest, lcom.ErrNoAuthorizationHeader
	}

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, http.StatusBadRequest, lcom.ErrNoBearerPrefix
	}

	userJwt := strings.TrimPrefix(authorizationHeader, "Bearer ")

	mapClaims, err := VerifyJWT(userJwt)
	if err != nil {
		return nil, http.StatusUnauthorized, util.WrapErrors(err, lcom.ErrVerifyJWT)
	}

	return mapClaims, http.StatusOK, nil
}
