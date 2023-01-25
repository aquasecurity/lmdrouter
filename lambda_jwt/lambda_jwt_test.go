package lambda_jwt

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/jgroeneveld/trial/assert"
	"github.com/joho/godotenv"
	"github.com/seantcanavan/lambda_jwt_router/lambda_util"
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
		Audience:  lambda_util.GenerateRandomString(10),
		ExpiresAt: time.Now().Add(time.Hour * 30).Unix(),
		FirstName: lambda_util.GenerateRandomString(10),
		FullName:  lambda_util.GenerateRandomString(10),
		ID:        lambda_util.GenerateRandomString(10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    lambda_util.GenerateRandomString(10),
		Level:     lambda_util.GenerateRandomString(10),
		NotBefore: time.Now().Add(time.Hour * -1).Unix(),
		Subject:   lambda_util.GenerateRandomString(10),
		UserType:  lambda_util.GenerateRandomString(10),
	}

	extendedClaims := ExtendExpanded(expandedClaims)

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
		Audience:  lambda_util.GenerateRandomString(10),
		ExpiresAt: time.Now().Add(time.Hour * 30).Unix(),
		Id:        lambda_util.GenerateRandomString(10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    lambda_util.GenerateRandomString(10),
		NotBefore: time.Now().Add(time.Hour * -1).Unix(),
		Subject:   lambda_util.GenerateRandomString(10),
	}

	extendedClaims := ExtendStandard(standardClaims)

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
	t.Run("verify ExtractCustom returns an err when unmarshalling to invalid custom claims object", func(t *testing.T) {
		type badClaims struct {
			ExpiresAt int64 `json:"exp"`
		}
		extractCustomErr := ExtractCustom(jwt.MapClaims{
			"exp": lambda_util.GenerateRandomString(10), // exp should be an integer
		}, &badClaims{})

		assert.NotNil(t, extractCustomErr)
		assert.True(t, errors.Is(extractCustomErr, ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustom works when called with the correct parameters", func(t *testing.T) {
		customClaims := generateExpandedMapClaims()

		var expandedClaims ExpandedClaims
		err := ExtractCustom(customClaims, &expandedClaims)
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
	t.Run("verify ExtractStandard returns an err when unmarshalling to invalid standard claims object", func(t *testing.T) {
		extractCustomErr := ExtractStandard(jwt.MapClaims{
			"exp": lambda_util.GenerateRandomString(10), // exp should be an integer
		}, &jwt.StandardClaims{})

		assert.NotNil(t, extractCustomErr)
		assert.True(t, errors.Is(extractCustomErr, ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustom works when called with the correct parameters", func(t *testing.T) {
		customClaims := generateExpandedMapClaims()

		var standardClaims jwt.StandardClaims
		err := ExtractCustom(customClaims, &standardClaims)
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
		customClaims := generateExpandedMapClaims()
		signedJWT, err := Sign(customClaims)
		assert.Nil(t, err)
		assert.True(t, len(signedJWT) > 1)
		assert.True(t, strings.HasPrefix(signedJWT, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
	})
}

func TestVerifyJWT(t *testing.T) {
	t.Run("verify err when parsing invalid jwt", func(t *testing.T) {
		_, err := VerifyJWT(lambda_util.GenerateRandomString(10))
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrInvalidJWT))
	})
	t.Run("verify err when parsing expired token with valid jwt", func(t *testing.T) {
		customClaims := generateExpandedMapClaims()
		customClaims["exp"] = time.Now().Add(time.Hour * -10)

		expiredJWT, signErr := Sign(customClaims)
		assert.Nil(t, signErr)

		_, err := VerifyJWT(expiredJWT)
		assert.NotNil(t, err)
		assert.True(t, errors.Is(err, ErrInvalidJWT))
	})
}

func generateExpandedMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		AudienceKey:  lambda_util.GenerateRandomString(10),
		ExpiresAtKey: time.Now().Add(time.Hour * 30).Unix(),
		FirstNameKey: lambda_util.GenerateRandomString(10),
		FullNameKey:  lambda_util.GenerateRandomString(10),
		IDKey:        lambda_util.GenerateRandomString(10),
		IssuedAtKey:  time.Now().Unix(),
		IssuerKey:    lambda_util.GenerateRandomString(10),
		LevelKey:     lambda_util.GenerateRandomString(10),
		NotBeforeKey: time.Now().Add(time.Hour * -1).Unix(),
		SubjectKey:   lambda_util.GenerateRandomString(10),
		UserTypeKey:  lambda_util.GenerateRandomString(10),
	}
}

func generateStandardMapClaims() jwt.MapClaims {
	return jwt.MapClaims{
		AudienceKey:  lambda_util.GenerateRandomString(10),
		ExpiresAtKey: time.Now().Add(time.Hour * 30).Unix(),
		IDKey:        lambda_util.GenerateRandomString(10),
		IssuedAtKey:  time.Now().Unix(),
		IssuerKey:    lambda_util.GenerateRandomString(10),
		NotBeforeKey: time.Now().Add(time.Hour * -1).Unix(),
		SubjectKey:   lambda_util.GenerateRandomString(10),
	}
}
