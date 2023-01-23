package lmdrouter

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"os"
)

var secret = os.Getenv("HMAC_SECRET")

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

func Sign(mapClaims *jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, mapClaims)

	// Sign and get the complete encoded token as a string using the secret
	encodedToken, err := token.SignedString(getBinarySecret())
	if err != nil {
		return "", err
	}

	return encodedToken, nil
}

// ExtractStandardClaims takes in the claims map that is used to create JWTs
// and returns the standard 7 values expected in all json web tokens
func ExtractStandardClaims(mapClaims *jwt.MapClaims, standardClaims *jwt.StandardClaims) error {
	jsonBytes, err := json.Marshal(mapClaims)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, standardClaims)
	if err != nil {
		return err
	}

	return nil
}

// ExtractCustomClaims takes in a claims map that is used to create JWTs
// and returns a generic interface value that you can use to convert
func ExtractCustomClaims(claims *jwt.MapClaims, val any) error {
	jsonBytes, err := json.Marshal(claims)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonBytes, &val)
	if err != nil {
		return err
	}

	return nil
}

// VerifyJWT accepts the user JWT from the Authorization header
// and returns the MapClaims OR a http status code and error set
func VerifyJWT(userJWT string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(userJWT, keyFunc)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("unable to parse MapClaims")
}

func getBinarySecret() []byte {
	data, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatalf("cannot decode the secret")
	}

	return data
}

func keyFunc(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return getBinarySecret(), nil
}