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

// ExtendExpandedClaims returns an instance of jwt.MapClaims which you can freely extend
// with your own custom fields. It uses ExpandedClaims as the base struct to start with.
func ExtendExpandedClaims(claims ExpandedClaims) jwt.MapClaims {
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

// ExtendStandardClaims returns an instance of jwt.MapClaims which you can freely extend
// with your own custom fields. It uses jwt.StandardClaims as the base struct to start with.
func ExtendStandardClaims(claims jwt.StandardClaims) jwt.MapClaims {
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

// ExtractCustomClaims takes in a claims map that is used to create JWTs
// and returns a generic interface value that you can use to convert
func ExtractCustomClaims(mapClaims jwt.MapClaims, val any) error {
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

// ExtractStandardClaims takes in the claims map that is used to create JWTs
// and returns the standard 7 values expected in all json web tokens
func ExtractStandardClaims(mapClaims jwt.MapClaims, standardClaims *jwt.StandardClaims) error {
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
// and returns the MapClaims OR a http status code and error set
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
