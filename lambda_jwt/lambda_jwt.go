// Package lambda_jwt appends critical libraries necessary for using JWTs (Json Web Tokens)
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
package lambda_jwt

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt"
	"log"
	"os"
)

// Use these const values to populate your own custom claim values
const (
	AudienceKey  = "aud"
	ExpiresAtKey = "exp"
	FirstNameKey = "firstName"
	FullNameKey  = "fullName"
	IDKey        = "jti"
	IssuedAtKey  = "iat"
	IssuerKey    = "iss"
	LevelKey     = "level"
	NotBeforeKey = "nbf"
	SubjectKey   = "sub"
	UserTypeKey  = "userType"
)

var ErrBadClaimsObject = errors.New("lambda_jwt_router: the provided object to extract claims into is not compatible with the default claim set and its types")
var ErrUnableToSignToken = errors.New("lambda_jwt_router: the provided claims were unable to be signed")
var ErrInvalidJWT = errors.New("lambda_jwt_router: the provided JWT is invalid")
var ErrInvalidToken = errors.New("lambda_jwt_router: the provided jwt was unable to be parsed into a token")
var ErrInvalidTokenClaims = errors.New("lambda_jwt_router: the provided jwt was unable to be parsed for map claims")
var ErrUnsupportedSigningMethod = errors.New("lambda_jwt_router:the provided signing method is unsupported. HMAC only allowed")

type ExpandedClaims struct {
	Audience  string `json:"aud"`
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
		AudienceKey:  claims.Audience,
		ExpiresAtKey: claims.ExpiresAt,
		FirstNameKey: claims.FirstName,
		FullNameKey:  claims.FullName,
		IDKey:        claims.ID,
		IssuedAtKey:  claims.IssuedAt,
		IssuerKey:    claims.Issuer,
		LevelKey:     claims.Level,
		NotBeforeKey: claims.NotBefore,
		SubjectKey:   claims.Subject,
		UserTypeKey:  claims.UserType,
	}
}

// ExtendStandard returns an instance of jwt.MapClaims which you can freely extend
// with your own custom fields. It uses jwt.StandardClaims as the base struct to start with
// and returns a jwt.MapClaims which is just a wrapper for a map so you can add as many
// custom fields as you would like while still getting the 7 standard JWT fields.
func ExtendStandard(claims jwt.StandardClaims) jwt.MapClaims {
	return jwt.MapClaims{
		AudienceKey:  claims.Audience,
		ExpiresAtKey: claims.ExpiresAt,
		IDKey:        claims.Id,
		IssuedAtKey:  claims.IssuedAt,
		IssuerKey:    claims.Issuer,
		NotBeforeKey: claims.NotBefore,
		SubjectKey:   claims.Subject,
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
		return err
	}

	err = json.Unmarshal(jsonBytes, &val)
	if err != nil {
		return ErrBadClaimsObject
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
		return err
	}

	err = json.Unmarshal(jsonBytes, standardClaims)
	if err != nil {
		return ErrBadClaimsObject
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
		return "", ErrUnableToSignToken
	}

	return encodedToken, nil
}

// VerifyJWT accepts the user JWT from the Authorization header
// and returns the MapClaims or nil and an error set.
func VerifyJWT(userJWT string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(userJWT, keyFunc)
	if err != nil {
		return nil, ErrInvalidJWT
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidTokenClaims
}

func getBinarySecret() []byte {
	secret := os.Getenv("HMAC_SECRET")
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
		return nil, ErrUnsupportedSigningMethod
	}

	return getBinarySecret(), nil
}
