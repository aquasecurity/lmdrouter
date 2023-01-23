package lmdrouter

import (
	"github.com/golang-jwt/jwt"
	"github.com/jgroeneveld/trial/assert"
	"math/rand"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestInitializeMapClaimsAndExtractStandardClaims(t *testing.T) {
	t.Run("recover full standard claims", func(t *testing.T) {
		claims, err := InitializeMapClaims(
			GenerateRandomString(10),
			20000,
			GenerateRandomString(10),
			20000,
			GenerateRandomString(10),
			20000,
			GenerateRandomString(10),
		)
		assert.Nil(t, err)

		standardClaims, err := ExtractStandardClaims(claims)
		assert.Nil(t, err)

		assert.Equal(t, standardClaims.Audience, claims[AudienceKey])
		assert.Equal(t, standardClaims.ExpiresAt, claims[ExpiresAtKey])
		assert.Equal(t, standardClaims.Id, claims[IDKey])
		assert.Equal(t, standardClaims.IssuedAt, claims[IssuedAtKey])
		assert.Equal(t, standardClaims.Issuer, claims[IssuerKey])
		assert.Equal(t, standardClaims.NotBefore, claims[NotBeforeKey])
		assert.Equal(t, standardClaims.Subject, claims[SubjectKey])
	})
}

func TestGenerateJWTAndVerifyJWT(t *testing.T) {
	t.Run("encode standard jwt inside map claims", func(t *testing.T) {
		claims, err := InitializeMapClaims(
			GenerateRandomString(10),
			time.Now().Add(time.Hour*1).Unix(), // expiresAt
			GenerateRandomString(10),
			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
			GenerateRandomString(10),
			time.Now().Unix()-50000, // notBefore
			GenerateRandomString(10),
		)
		assert.Nil(t, err)
		jwt, httpStatus, err := GenerateJWT(claims)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.Nil(t, err)
		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
	})

	t.Run("encode standard jwt with custom claims", func(t *testing.T) {
		claims, err := InitializeMapClaims(
			GenerateRandomString(10),
			time.Now().Add(time.Hour*1).Unix(), // expiresAt
			GenerateRandomString(10),
			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
			GenerateRandomString(10),
			time.Now().Unix()-50000, // notBefore
			GenerateRandomString(10),
		)
		assert.Nil(t, err)

		claims["custom"] = "value"
		claims["author"] = "sean"

		jwt, httpStatus, err := GenerateJWT(claims)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
	})

	t.Run("verify audience", func(t *testing.T) {
		audience := GenerateRandomString(10)
		issuer := GenerateRandomString(10)
		mapClaims, err := InitializeMapClaims(
			audience,
			time.Now().Add(time.Hour*1).Unix(),  // expiresAt
			GenerateRandomString(10),            // id
			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
			issuer,
			time.Now().Unix()-50000,  // notBefore
			GenerateRandomString(10), // subject
		)
		assert.Nil(t, err)

		mapClaims["custom"] = "value"
		mapClaims["author"] = "sean"

		jwt, httpStatus, err := GenerateJWT(mapClaims)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))

		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.NotNil(t, retrievedClaims)

		audienceVerified := retrievedClaims.VerifyAudience(audience, true)
		assert.True(t, audienceVerified)

		issuerVerified := retrievedClaims.VerifyIssuer(issuer, true)
		assert.True(t, issuerVerified)
	})

	t.Run("jwt is expired", func(t *testing.T) {
		claims, err := InitializeMapClaims(
			GenerateRandomString(10),
			time.Now().Add(time.Hour*-1).Unix(), // expiresAt
			GenerateRandomString(10),
			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
			GenerateRandomString(10),
			time.Now().Unix()-50000, // notBefore
			GenerateRandomString(10),
		)
		assert.Nil(t, err)

		jwt, httpStatus, err := GenerateJWT(claims)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))

		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
		assert.True(t, len(retrievedClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "Token is expired")
	})

	t.Run("issued in the future", func(t *testing.T) {
		claims, err := InitializeMapClaims(
			GenerateRandomString(10),
			time.Now().Add(time.Hour*2).Unix(), // expiresAt
			GenerateRandomString(10),
			time.Now().Add(time.Hour*1).Unix(), // issuedAt
			GenerateRandomString(10),
			time.Now().Unix()-50000, // notBefore
			GenerateRandomString(10),
		)
		assert.Nil(t, err)

		jwt, httpStatus, err := GenerateJWT(claims)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))

		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
		assert.True(t, len(retrievedClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "Token used before issued")
	})

	t.Run("used before allowed", func(t *testing.T) {
		claims, err := InitializeMapClaims(
			GenerateRandomString(10),
			time.Now().Add(time.Hour*2).Unix(), // expiresAt
			GenerateRandomString(10),
			time.Now().Add(time.Hour*1).Unix(), // issuedAt
			GenerateRandomString(10),
			time.Now().Add(time.Hour*1).Unix(), // notBefore
			GenerateRandomString(10),
		)
		assert.Nil(t, err)

		jwt, httpStatus, err := GenerateJWT(claims)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))

		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
		assert.True(t, len(retrievedClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, err)
		assert.Equal(t, err.Error(), "Token is not valid yet")
	})
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenerateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func GenerateCustomMapClaims() *jwt.MapClaims {
	return &jwt.MapClaims{
		AudienceKey:  GenerateRandomString(10),
		ExpiresAtKey: GenerateRandomString(10),
		FirstNameKey: GenerateRandomString(10),
		FullNameKey:  GenerateRandomString(10),
		IDKey:        GenerateRandomString(10),
		IssuedAtKey:  time.Now().Unix(),
		IssuerKey:    GenerateRandomString(10),
		LevelKey:     GenerateRandomString(10),
		NotBeforeKey: time.Now().Add(time.Hour * -1).Unix(),
		SubjectKey:   GenerateRandomString(10),
		UserTypeKey:  GenerateRandomString(10),
	}
}
