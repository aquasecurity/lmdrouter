package jwt_auth

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/jgroeneveld/trial/assert"
	"github.com/joho/godotenv"
	"github.com/seantcanavan/lmdrouter/utils"
	"log"
	"strings"
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

func TestExtendExpandedClaims(t *testing.T) {
	expandedClaims := ExpandedClaims{
		Audience:  utils.GenerateRandomString(10),
		ExpiresAt: time.Now().Add(time.Hour * 30).Unix(),
		FirstName: utils.GenerateRandomString(10),
		FullName:  utils.GenerateRandomString(10),
		ID:        utils.GenerateRandomString(10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    utils.GenerateRandomString(10),
		Level:     utils.GenerateRandomString(10),
		NotBefore: time.Now().Add(time.Hour * -1).Unix(),
		Subject:   utils.GenerateRandomString(10),
		UserType:  utils.GenerateRandomString(10),
	}

	extendedClaims := ExtendExpandedClaims(expandedClaims)

	extendedClaims["hi"] = "sean"
	extendedClaims["hello"] = "there"
	extendedClaims["number"] = 34

	t.Run("verify sign and verify expanded and custom fields in claims", func(t *testing.T) {
		signedJWT, signErr := Sign(extendedClaims)
		assert.Nil(t, signErr)

		retrievedClaims, verifyErr := VerifyJWT(signedJWT)
		assert.Nil(t, verifyErr)

		// verify the expanded claims values first
		assert.Equal(t, retrievedClaims[AudienceKey], expandedClaims.Audience)
		assert.Equal(t, retrievedClaims[ExpiresAtKey], float64(expandedClaims.ExpiresAt))
		assert.Equal(t, retrievedClaims[FirstNameKey], expandedClaims.FirstName)
		assert.Equal(t, retrievedClaims[IDKey], expandedClaims.ID)
		assert.Equal(t, retrievedClaims[IssuedAtKey], float64(expandedClaims.IssuedAt))
		assert.Equal(t, retrievedClaims[IssuerKey], expandedClaims.Issuer)
		assert.Equal(t, retrievedClaims[LevelKey], expandedClaims.Level)
		assert.Equal(t, retrievedClaims[NotBeforeKey], float64(expandedClaims.NotBefore))
		assert.Equal(t, retrievedClaims[SubjectKey], expandedClaims.Subject)
		assert.Equal(t, retrievedClaims[UserTypeKey], expandedClaims.UserType)

		// verify the custom claim values second
		assert.Equal(t, retrievedClaims["hi"], "sean")
		assert.Equal(t, retrievedClaims["hello"], "there")
		assert.Equal(t, retrievedClaims["number"], float64(34))
	})
}

func TestExtendStandardClaims(t *testing.T) {
	standardClaims := jwt.StandardClaims{
		Audience:  utils.GenerateRandomString(10),
		ExpiresAt: time.Now().Add(time.Hour * 30).Unix(),
		Id:        utils.GenerateRandomString(10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    utils.GenerateRandomString(10),
		NotBefore: time.Now().Add(time.Hour * -1).Unix(),
		Subject:   utils.GenerateRandomString(10),
	}

	extendedClaims := ExtendStandardClaims(standardClaims)

	extendedClaims["hi"] = "sean"
	extendedClaims["hello"] = "there"
	extendedClaims["number"] = 34

	t.Run("verify sign and verify standard and custom fields in claims", func(t *testing.T) {
		signedJWT, signErr := Sign(extendedClaims)
		assert.Nil(t, signErr)

		retrievedClaims, verifyErr := VerifyJWT(signedJWT)
		assert.Nil(t, verifyErr)

		// verify the expanded claims values first
		assert.Equal(t, retrievedClaims[AudienceKey], standardClaims.Audience)
		assert.Equal(t, retrievedClaims[ExpiresAtKey], float64(standardClaims.ExpiresAt))
		assert.Equal(t, retrievedClaims[IDKey], standardClaims.Id)
		assert.Equal(t, retrievedClaims[IssuedAtKey], float64(standardClaims.IssuedAt))
		assert.Equal(t, retrievedClaims[IssuerKey], standardClaims.Issuer)
		assert.Equal(t, retrievedClaims[NotBeforeKey], float64(standardClaims.NotBefore))
		assert.Equal(t, retrievedClaims[SubjectKey], standardClaims.Subject)

		// verify the custom claim values second
		assert.Equal(t, retrievedClaims["hi"], "sean")
		assert.Equal(t, retrievedClaims["hello"], "there")
		assert.Equal(t, retrievedClaims["number"], float64(34))
	})
}

func TestExtractCustomClaims(t *testing.T) {
	t.Run("verify ExtractCustomClaims returns an err when unmarshalling to invalid custom claims object", func(t *testing.T) {
		type badClaims struct {
			ExpiresAt int64 `json:"exp"`
		}
		extractCustomErr := ExtractCustomClaims(jwt.MapClaims{
			"exp": utils.GenerateRandomString(10), // exp should be an integer
		}, &badClaims{})

		assert.NotNil(t, extractCustomErr)
		assert.True(t, errors.Is(extractCustomErr, ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustomClaims works when called with the correct parameters", func(t *testing.T) {
		customClaims := GenerateExtendedMapClaims()

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
			"exp": utils.GenerateRandomString(10), // exp should be an integer
		}, jwt.StandardClaims{})

		assert.NotNil(t, extractCustomErr)
		assert.True(t, errors.Is(extractCustomErr, ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustomClaims works when called with the correct parameters", func(t *testing.T) {
		customClaims := GenerateExtendedMapClaims()

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
		customClaims := GenerateExtendedMapClaims()
		signedJWT, err := Sign(customClaims)
		assert.Nil(t, err)
		assert.True(t, len(signedJWT) > 1)
		assert.True(t, strings.HasPrefix(signedJWT, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
	})
}

func TestVerifyJWT(t *testing.T) {
	t.Run("verify err when parsing invalid jwt", func(t *testing.T) {
		_, err := VerifyJWT(utils.GenerateRandomString(10))
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrInvalidJWT))
	})
	t.Run("verify err when parsing expired token with valid jwt", func(t *testing.T) {
		customClaims := GenerateExtendedMapClaims()
		customClaims["exp"] = time.Now().Add(time.Hour * -10)

		expiredJWT, signErr := Sign(customClaims)
		assert.Nil(t, signErr)

		_, err := VerifyJWT(expiredJWT)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrInvalidJWT))
	})
}

func GenerateExtendedMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		AudienceKey:  utils.GenerateRandomString(10),
		ExpiresAtKey: time.Now().Add(time.Hour * 30).Unix(),
		FirstNameKey: utils.GenerateRandomString(10),
		FullNameKey:  utils.GenerateRandomString(10),
		IDKey:        utils.GenerateRandomString(10),
		IssuedAtKey:  time.Now().Unix(),
		IssuerKey:    utils.GenerateRandomString(10),
		LevelKey:     utils.GenerateRandomString(10),
		NotBeforeKey: time.Now().Add(time.Hour * -1).Unix(),
		SubjectKey:   utils.GenerateRandomString(10),
		UserTypeKey:  utils.GenerateRandomString(10),
	}
}

func GenerateStandardMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		AudienceKey:  utils.GenerateRandomString(10),
		ExpiresAtKey: time.Now().Add(time.Hour * 30).Unix(),
		IDKey:        utils.GenerateRandomString(10),
		IssuedAtKey:  time.Now().Unix(),
		IssuerKey:    utils.GenerateRandomString(10),
		NotBeforeKey: time.Now().Add(time.Hour * -1).Unix(),
		SubjectKey:   utils.GenerateRandomString(10),
	}
}
