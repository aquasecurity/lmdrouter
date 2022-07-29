package lmdrouter

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"os"
)

var secret = os.Getenv("HMAC_SECRET")

const (
	AudienceKey  = "aud"
	ExpiresAtKey = "exp"
	IDKey        = "jti"
	IssuedAtKey  = "iat"
	IssuerKey    = "iss"
	NotBeforeKey = "nbf"
	SubjectKey   = "sub"
)

func GenerateJWT(claims jwt.MapClaims) (string, int, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	// Sign and get the complete encoded token as a string using the secret
	encodedToken, err := token.SignedString(getBinarySecret())
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	return encodedToken, http.StatusOK, nil
}

// InitializeMapClaims embeds all the necessary values for a StandardClaims inside of custom MapClaims.
// The returned MapClaims can be used to insert any custom values that you need for your own purpose.
func InitializeMapClaims(audience string, expiresAt int64, id string, issuedAt int64, issuer string, notBefore int64, subject string) (jwt.MapClaims, error) {
	return verifyMapClaims(jwt.MapClaims{
		AudienceKey:  audience,
		ExpiresAtKey: expiresAt,
		IDKey:        id,
		IssuedAtKey:  issuedAt,
		IssuerKey:    issuer,
		NotBeforeKey: notBefore,
		SubjectKey:   subject,
	})
}

func ExtractStandardClaims(mapClaims jwt.MapClaims) (*jwt.StandardClaims, error) {
	jsonBytes, err := json.Marshal(mapClaims)
	if err != nil {
		return nil, err
	}

	var standardClaims jwt.StandardClaims
	err = json.Unmarshal(jsonBytes, &standardClaims)
	if err != nil {
		return nil, err
	}

	return &standardClaims, nil
}

func verifyMapClaims(claims jwt.MapClaims) (jwt.MapClaims, error) {
	if claims["aud"].(string) == "" {
		return nil, fmt.Errorf("aud (audience) is required")
	}

	if claims["exp"].(int64) == 0 {
		return nil, fmt.Errorf("exp (expiresAt) is required")
	}

	if claims["jti"].(string) == "" {
		return nil, fmt.Errorf("jti (id) is required")
	}

	if claims["iat"].(int64) == 0 {
		return nil, fmt.Errorf("iat (issuedAt) is required")
	}

	if claims["iss"].(string) == "" {
		return nil, fmt.Errorf("iss (issuer) is required")
	}

	if claims["nbf"].(int64) == 0 {
		return nil, fmt.Errorf("nbf (notBefore) is required")
	}

	if claims["sub"].(string) == "" {
		return nil, fmt.Errorf("sub (subject) is required")
	}

	return claims, nil
}

// VerifyJWT accepts the user JWT from the Authorization header
// and returns the MapClaims OR a http status code and error set
func VerifyJWT(userJWT string) (jwt.MapClaims, int, error) {
	token, err := jwt.Parse(userJWT, keyFunc)
	if err != nil {
		return nil, http.StatusUnauthorized, err
	}

	if !token.Valid {
		return nil, http.StatusUnauthorized, errors.New("token is not valid")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, http.StatusOK, nil
	}

	return nil, http.StatusInternalServerError, errors.New("unable to parse MapClaims")
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
