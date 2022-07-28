package lmdrouter

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type AdminClaim struct {
	ExpirationDate time.Time
	FullName       string
	ID             string
	Level          string
}

func GenerateJwt(adminFullName, adminID, adminLevel string, expirationDate *time.Time) (string, int, error) {
	if adminFullName == "" {
		return "", 400, fmt.Errorf("unable to generate jwt with empty adminFullName %s", adminFullName)
	}

	if adminID == "" {
		return "", 400, fmt.Errorf("unable to generate jwt with empty adminID %s", adminID)
	}

	if adminLevel == "" {
		return "", 400, fmt.Errorf("unable to generate jwt with empty adminLevel %s", adminLevel)
	}

	if expirationDate.IsZero() {
		return "", 400, fmt.Errorf("unable to generate jwt with empty expirationDate %+v", expirationDate)
	}

	if expirationDate.Before(time.Now()) {
		return "", 400, fmt.Errorf("unable to generate jwt with expiration before now %+v", expirationDate)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"adminFullName":  adminFullName,
		"adminID":        adminID,
		"adminLevel":     adminLevel,
		"expirationDate": expirationDate,
	})

	// Sign and get the complete encoded token as a string using the secret
	encodedToken, err := token.SignedString(getBinarySecret())
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	return encodedToken, http.StatusOK, nil
}

func JwtAuth(headers map[string]string) (*AdminClaim, int, error) {
	authorizationHeader := headers["Authorization"]
	if authorizationHeader == "" {
		return nil, http.StatusForbidden, fmt.Errorf("missing Authorization header value")
	}

	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, http.StatusForbidden, fmt.Errorf("missing 'Bearer ' prefix for Authorization header")
	}

	userJwt := strings.TrimPrefix(authorizationHeader, "Bearer ")

	adminClaim, httpStatus, err := Verify(userJwt)
	if err != nil {
		return nil, httpStatus, err
	}

	return adminClaim, http.StatusOK, nil
}

// Verify accepts the user JWT from the Authorization header
// and returns the adminID, adminLevel, true/false for success
// and an error code
func Verify(userJwt string) (*AdminClaim, int, error) {
	token, err := jwt.Parse(userJwt, KeyFunc)

	if err != nil {
		log.Printf("Error with JWT %+v", err)
		return nil, http.StatusInternalServerError, err
	}

	if !token.Valid {
		return nil, http.StatusForbidden, errors.New("token is not valid")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		adminFullName := claims["adminFullName"].(string)
		adminID := claims["adminID"].(string)
		adminLevel := claims["adminLevel"].(string)
		expirationDate := claims["expirationDate"].(string)

		expirationDateAsTime, err := time.Parse(time.RFC3339, expirationDate)
		if err != nil {
			return nil, 500, err
		}

		if adminID == "" {
			return nil, http.StatusForbidden, errors.New("adminID cannot be empty")
		}

		if adminLevel == "" {
			return nil, http.StatusForbidden, errors.New("adminLevel cannot be empty")
		}

		if expirationDateAsTime.IsZero() {
			return nil, http.StatusUnauthorized, errors.New("invalid expiration date")
		}

		if time.Now().After(expirationDateAsTime) {
			return nil, http.StatusUnauthorized, errors.New("jwt is expired")
		}

		return &AdminClaim{
			ExpirationDate: expirationDateAsTime,
			FullName:       adminFullName,
			ID:             adminID,
			Level:          adminLevel,
		}, http.StatusOK, nil
	}

	return nil, http.StatusInternalServerError, errors.New("unable to parse MapClaims")
}

var secret = os.Getenv("HMAC_SECRET")

func getBinarySecret() []byte {
	data, err := hex.DecodeString(secret)
	if err != nil {
		log.Fatalf("cannot decode the secret")
	}

	return data
}

func KeyFunc(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	}

	return getBinarySecret(), nil
}
