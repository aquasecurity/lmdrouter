package jwt

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/jgroeneveld/trial/assert"
	"log"
	"math/rand"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
}

func setup() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Unable to load .env file: %s", err)
	}
}

func TestExtractCustomClaims(t *testing.T) {
	t.Run("verify ExtractCustomClaims returns an err when unmarshalling to invalid custom claims object", func(t *testing.T) {
		type badClaims struct {
			ExpiresAt int64 `json:"exp"`
		}
		extractCustomErr := ExtractCustomClaims(jwt.MapClaims{
			"exp": GenerateRandomString(10), // exp should be an integer
		}, &badClaims{})

		assert.NotNil(t, extractCustomErr)
		assert.True(t, errors.Is(extractCustomErr, ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustomClaims works when called with the correct parameters", func(t *testing.T) {
		customClaims := GenerateCustomMapClaims()

		var expandedClaims ExpandedClaims
		err := ExtractCustomClaims(customClaims, &expandedClaims)
		assert.Nil(t, err)

		assert.Equal(t, customClaims[AudienceKey], expandedClaims.Audience)
		assert.Equal(t, customClaims[ExpiresAtKey], expandedClaims.ExpiresAt)
		assert.Equal(t, customClaims[FirstNameKey], expandedClaims.FirstName)
		assert.Equal(t, customClaims[FullNameKey], expandedClaims.FullName)
		assert.Equal(t, customClaims[IDKey], expandedClaims.ID)
		assert.Equal(t, customClaims[IssuedAtKey], expandedClaims.IssuedAt)
		assert.Equal(t, customClaims[IssuerKey], expandedClaims.Issuer)
		assert.Equal(t, customClaims[LevelKey], expandedClaims.Level)
		assert.Equal(t, customClaims[NotBeforeKey], expandedClaims.NotBefore)
		assert.Equal(t, customClaims[SubjectKey], expandedClaims.Subject)
		assert.Equal(t, customClaims[UserTypeKey], expandedClaims.UserType)
	})
}

func TestExtractStandardClaims(t *testing.T) {
	t.Run("verify ExtractStandardClaims returns an err when unmarshalling to invalid standard claims object", func(t *testing.T) {
		extractCustomErr := ExtractStandardClaims(jwt.MapClaims{
			"exp": GenerateRandomString(10), // exp should be an integer
		}, jwt.StandardClaims{})

		assert.NotNil(t, extractCustomErr)
		assert.True(t, errors.Is(extractCustomErr, ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustomClaims works when called with the correct parameters", func(t *testing.T) {
		customClaims := GenerateCustomMapClaims()

		var standardClaims jwt.StandardClaims
		err := ExtractCustomClaims(customClaims, &standardClaims)
		assert.Nil(t, err)

		assert.Equal(t, customClaims[AudienceKey], standardClaims.Audience)
		assert.Equal(t, customClaims[ExpiresAtKey], standardClaims.ExpiresAt)
		assert.Equal(t, customClaims[IssuedAtKey], standardClaims.IssuedAt)
		assert.Equal(t, customClaims[IssuerKey], standardClaims.Issuer)
		assert.Equal(t, customClaims[NotBeforeKey], standardClaims.NotBefore)
		assert.Equal(t, customClaims[SubjectKey], standardClaims.Subject)
	})
}

func TestSign(t *testing.T) {
	t.Run("verify signed jwt secret with valid standard claim", func(t *testing.T) {
		fmt.Println(fmt.Sprintf("secret is [%s]", secret))
		customClaims := GenerateCustomMapClaims()
		signedJWT, err := Sign(customClaims)
		assert.Nil(t, err)
		assert.True(t, len(signedJWT) > 1)
	})
}

//func TestInitializeMapClaimsAndExtractStandardClaims(t *testing.T) {
//	t.Run("recover full standard claims", func(t *testing.T) {
//		claims, err := InitializeMapClaims(
//			GenerateRandomString(10),
//			20000,
//			GenerateRandomString(10),
//			20000,
//			GenerateRandomString(10),
//			20000,
//			GenerateRandomString(10),
//		)
//		assert.Nil(t, err)
//
//		standardClaims, err := ExtractStandardClaims(claims)
//		assert.Nil(t, err)
//
//		assert.Equal(t, standardClaims.Audience, claims[AudienceKey])
//		assert.Equal(t, standardClaims.ExpiresAt, claims[ExpiresAtKey])
//		assert.Equal(t, standardClaims.Id, claims[IDKey])
//		assert.Equal(t, standardClaims.IssuedAt, claims[IssuedAtKey])
//		assert.Equal(t, standardClaims.Issuer, claims[IssuerKey])
//		assert.Equal(t, standardClaims.NotBefore, claims[NotBeforeKey])
//		assert.Equal(t, standardClaims.Subject, claims[SubjectKey])
//	})
//}
//
//func TestGenerateJWTAndVerifyJWT(t *testing.T) {
//	t.Run("encode standard jwt inside map claims", func(t *testing.T) {
//		claims, err := InitializeMapClaims(
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*1).Unix(), // expiresAt
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
//			GenerateRandomString(10),
//			time.Now().Unix()-50000, // notBefore
//			GenerateRandomString(10),
//		)
//		assert.Nil(t, err)
//		jwt, httpStatus, err := GenerateJWT(claims)
//		assert.Equal(t, http.StatusOK, httpStatus)
//		assert.Nil(t, err)
//		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
//	})
//
//	t.Run("encode standard jwt with custom claims", func(t *testing.T) {
//		claims, err := InitializeMapClaims(
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*1).Unix(), // expiresAt
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
//			GenerateRandomString(10),
//			time.Now().Unix()-50000, // notBefore
//			GenerateRandomString(10),
//		)
//		assert.Nil(t, err)
//
//		claims["custom"] = "value"
//		claims["author"] = "sean"
//
//		jwt, httpStatus, err := GenerateJWT(claims)
//		assert.Nil(t, err)
//		assert.Equal(t, http.StatusOK, httpStatus)
//		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
//	})
//
//	t.Run("verify audience", func(t *testing.T) {
//		audience := GenerateRandomString(10)
//		issuer := GenerateRandomString(10)
//		mapClaims, err := InitializeMapClaims(
//			audience,
//			time.Now().Add(time.Hour*1).Unix(),  // expiresAt
//			GenerateRandomString(10),            // id
//			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
//			issuer,
//			time.Now().Unix()-50000,  // notBefore
//			GenerateRandomString(10), // subject
//		)
//		assert.Nil(t, err)
//
//		mapClaims["custom"] = "value"
//		mapClaims["author"] = "sean"
//
//		jwt, httpStatus, err := GenerateJWT(mapClaims)
//		assert.Nil(t, err)
//		assert.Equal(t, http.StatusOK, httpStatus)
//		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
//
//		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
//		assert.Nil(t, err)
//		assert.Equal(t, http.StatusOK, httpStatus)
//		assert.NotNil(t, retrievedClaims)
//
//		audienceVerified := retrievedClaims.VerifyAudience(audience, true)
//		assert.True(t, audienceVerified)
//
//		issuerVerified := retrievedClaims.VerifyIssuer(issuer, true)
//		assert.True(t, issuerVerified)
//	})
//
//	t.Run("jwt is expired", func(t *testing.T) {
//		claims, err := InitializeMapClaims(
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*-1).Unix(), // expiresAt
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*-1).Unix(), // issuedAt
//			GenerateRandomString(10),
//			time.Now().Unix()-50000, // notBefore
//			GenerateRandomString(10),
//		)
//		assert.Nil(t, err)
//
//		jwt, httpStatus, err := GenerateJWT(claims)
//		assert.Nil(t, err)
//		assert.Equal(t, http.StatusOK, httpStatus)
//		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
//
//		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
//		assert.True(t, len(retrievedClaims) == 0)
//		assert.Equal(t, http.StatusUnauthorized, httpStatus)
//		assert.NotNil(t, err)
//		assert.Equal(t, err.Error(), "Token is expired")
//	})
//
//	t.Run("issued in the future", func(t *testing.T) {
//		claims, err := InitializeMapClaims(
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*2).Unix(), // expiresAt
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*1).Unix(), // issuedAt
//			GenerateRandomString(10),
//			time.Now().Unix()-50000, // notBefore
//			GenerateRandomString(10),
//		)
//		assert.Nil(t, err)
//
//		jwt, httpStatus, err := GenerateJWT(claims)
//		assert.Nil(t, err)
//		assert.Equal(t, http.StatusOK, httpStatus)
//		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
//
//		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
//		assert.True(t, len(retrievedClaims) == 0)
//		assert.Equal(t, http.StatusUnauthorized, httpStatus)
//		assert.NotNil(t, err)
//		assert.Equal(t, err.Error(), "Token used before issued")
//	})
//
//	t.Run("used before allowed", func(t *testing.T) {
//		claims, err := InitializeMapClaims(
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*2).Unix(), // expiresAt
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*1).Unix(), // issuedAt
//			GenerateRandomString(10),
//			time.Now().Add(time.Hour*1).Unix(), // notBefore
//			GenerateRandomString(10),
//		)
//		assert.Nil(t, err)
//
//		jwt, httpStatus, err := GenerateJWT(claims)
//		assert.Nil(t, err)
//		assert.Equal(t, http.StatusOK, httpStatus)
//		assert.True(t, strings.HasPrefix(jwt, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
//
//		retrievedClaims, httpStatus, err := VerifyJWT(jwt)
//		assert.True(t, len(retrievedClaims) == 0)
//		assert.Equal(t, http.StatusUnauthorized, httpStatus)
//		assert.NotNil(t, err)
//		assert.Equal(t, err.Error(), "Token is not valid yet")
//	})
//}

func GenerateRandomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}

func GenerateCustomMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		AudienceKey:  GenerateRandomString(10),
		ExpiresAtKey: time.Now().Add(time.Hour * 30).Unix(),
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
