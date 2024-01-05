package ljwt

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/seantcanavan/lambda_jwt_router/lcom"
	"github.com/stretchr/testify/require"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	setup()
	m.Run()
}

func setup() {
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatalf("Unable to load .env file: %s", err)
	}
}

func TestExtendExpandedClaims(t *testing.T) {
	expandedClaims := ExpandedClaims{
		Audience:  util.GenerateRandomString(10),
		Email:     util.GenerateRandomString(10),
		ExpiresAt: time.Now().Add(time.Hour * 30).Unix(),
		FirstName: util.GenerateRandomString(10),
		FullName:  util.GenerateRandomString(10),
		ID:        util.GenerateRandomString(10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    util.GenerateRandomString(10),
		Level:     util.GenerateRandomString(10),
		NotBefore: time.Now().Add(time.Hour * -1).Unix(),
		Subject:   util.GenerateRandomString(10),
		UserType:  util.GenerateRandomString(10),
	}

	extendedClaims := ExtendExpanded(expandedClaims)

	extendedClaims["hi"] = "sean"
	extendedClaims["hello"] = "there"
	extendedClaims["number"] = 34

	t.Run("verify sign and verify expanded and custom fields in claims", func(t *testing.T) {
		signedJWT, signErr := Sign(extendedClaims)
		require.Nil(t, signErr)

		retrievedClaims, verifyErr := VerifyJWT(signedJWT)
		require.Nil(t, verifyErr)

		// verify the expanded claims values first
		require.Equal(t, retrievedClaims[lcom.JWTClaimAudienceKey], expandedClaims.Audience)
		require.Equal(t, retrievedClaims[lcom.JWTClaimExpiresAtKey], float64(expandedClaims.ExpiresAt))
		require.Equal(t, retrievedClaims[lcom.JWTClaimFirstNameKey], expandedClaims.FirstName)
		require.Equal(t, retrievedClaims[lcom.JWTClaimIDKey], expandedClaims.ID)
		require.Equal(t, retrievedClaims[lcom.JWTClaimIssuedAtKey], float64(expandedClaims.IssuedAt))
		require.Equal(t, retrievedClaims[lcom.JWTClaimIssuerKey], expandedClaims.Issuer)
		require.Equal(t, retrievedClaims[lcom.JWTClaimLevelKey], expandedClaims.Level)
		require.Equal(t, retrievedClaims[lcom.JWTClaimNotBeforeKey], float64(expandedClaims.NotBefore))
		require.Equal(t, retrievedClaims[lcom.JWTClaimSubjectKey], expandedClaims.Subject)
		require.Equal(t, retrievedClaims[lcom.JWTClaimUserTypeKey], expandedClaims.UserType)

		// verify the custom claim values second
		require.Equal(t, retrievedClaims["hi"], "sean")
		require.Equal(t, retrievedClaims["hello"], "there")
		require.Equal(t, retrievedClaims["number"], float64(34))
	})
}

func TestExtractCustomClaims(t *testing.T) {
	t.Run("verify ExtractCustom returns an err when unmarshalling to invalid custom claims object", func(t *testing.T) {
		type badClaims struct {
			ExpiresAt int64 `json:"exp"`
		}
		extractCustomErr := ExtractCustom(jwt.MapClaims{
			"exp": util.GenerateRandomString(10), // exp should be an integer
		}, &badClaims{})

		require.NotNil(t, extractCustomErr)
		require.True(t, errors.Is(extractCustomErr, lcom.ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustom works when called with the correct parameters", func(t *testing.T) {
		customClaims := util.GenerateExpandedMapClaims()

		var expandedClaims ExpandedClaims
		err := ExtractCustom(customClaims, &expandedClaims)
		require.Nil(t, err)

		require.Equal(t, customClaims[lcom.JWTClaimAudienceKey], expandedClaims.Audience)
		require.Equal(t, customClaims[lcom.JWTClaimExpiresAtKey], expandedClaims.ExpiresAt)
		require.Equal(t, customClaims[lcom.JWTClaimEmailKey], expandedClaims.Email)
		require.Equal(t, customClaims[lcom.JWTClaimFirstNameKey], expandedClaims.FirstName)
		require.Equal(t, customClaims[lcom.JWTClaimFullNameKey], expandedClaims.FullName)
		require.Equal(t, customClaims[lcom.JWTClaimIDKey], expandedClaims.ID)
		require.Equal(t, customClaims[lcom.JWTClaimIssuedAtKey], expandedClaims.IssuedAt)
		require.Equal(t, customClaims[lcom.JWTClaimIssuerKey], expandedClaims.Issuer)
		require.Equal(t, customClaims[lcom.JWTClaimLevelKey], expandedClaims.Level)
		require.Equal(t, customClaims[lcom.JWTClaimNotBeforeKey], expandedClaims.NotBefore)
		require.Equal(t, customClaims[lcom.JWTClaimSubjectKey], expandedClaims.Subject)
		require.Equal(t, customClaims[lcom.JWTClaimUserTypeKey], expandedClaims.UserType)
	})
}

func TestExtractStandardClaims(t *testing.T) {
	t.Run("verify ExtractStandard returns an err when unmarshalling to invalid standard claims object", func(t *testing.T) {
		extractCustomErr := ExtractStandard(jwt.MapClaims{
			"exp": util.GenerateRandomString(10), // exp should be an integer
		}, &jwt.StandardClaims{})

		require.NotNil(t, extractCustomErr)
		require.True(t, errors.Is(extractCustomErr, lcom.ErrBadClaimsObject))
	})
	t.Run("verify ExtractCustom works when called with the correct parameters", func(t *testing.T) {
		customClaims := util.GenerateExpandedMapClaims()

		var standardClaims jwt.StandardClaims
		err := ExtractCustom(customClaims, &standardClaims)
		require.Nil(t, err)

		require.Equal(t, customClaims[lcom.JWTClaimAudienceKey], standardClaims.Audience)
		require.Equal(t, customClaims[lcom.JWTClaimExpiresAtKey], standardClaims.ExpiresAt)
		require.Equal(t, customClaims[lcom.JWTClaimIssuedAtKey], standardClaims.IssuedAt)
		require.Equal(t, customClaims[lcom.JWTClaimIssuerKey], standardClaims.Issuer)
		require.Equal(t, customClaims[lcom.JWTClaimNotBeforeKey], standardClaims.NotBefore)
		require.Equal(t, customClaims[lcom.JWTClaimSubjectKey], standardClaims.Subject)
	})
}

func TestSign(t *testing.T) {
	t.Run("verify signed jwt secret with valid standard claim", func(t *testing.T) {
		customClaims := util.GenerateExpandedMapClaims()
		signedJWT, err := Sign(customClaims)
		require.Nil(t, err)
		require.True(t, len(signedJWT) > 1)
		require.True(t, strings.HasPrefix(signedJWT, "eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9"))
	})
}

func TestVerifyJWT(t *testing.T) {
	t.Run("verify err when parsing invalid jwt", func(t *testing.T) {
		_, err := VerifyJWT(util.GenerateRandomString(10))
		require.NotNil(t, err)
		require.True(t, errors.Is(err, lcom.ErrInvalidJWT))
	})
	t.Run("verify err when parsing expired token with valid jwt", func(t *testing.T) {
		customClaims := util.GenerateExpandedMapClaims()
		customClaims["exp"] = time.Now().Add(time.Hour * -10)

		expiredJWT, signErr := Sign(customClaims)
		require.Nil(t, signErr)

		_, err := VerifyJWT(expiredJWT)
		require.NotNil(t, err)
		require.True(t, errors.Is(err, lcom.ErrInvalidJWT))
	})
}

func TestExtendStandardClaims(t *testing.T) {
	standardClaims := jwt.StandardClaims{
		Audience:  util.GenerateRandomString(10),
		ExpiresAt: time.Now().Add(time.Hour * 30).Unix(),
		Id:        util.GenerateRandomString(10),
		IssuedAt:  time.Now().Unix(),
		Issuer:    util.GenerateRandomString(10),
		NotBefore: time.Now().Add(time.Hour * -1).Unix(),
		Subject:   util.GenerateRandomString(10),
	}

	extendedClaims := ExtendStandard(standardClaims)

	extendedClaims["hi"] = "sean"
	extendedClaims["hello"] = "there"
	extendedClaims["number"] = 34

	t.Run("verify sign and verify standard and custom fields in claims", func(t *testing.T) {
		signedJWT, signErr := Sign(extendedClaims)
		require.Nil(t, signErr)

		retrievedClaims, verifyErr := VerifyJWT(signedJWT)
		require.Nil(t, verifyErr)

		// verify the expanded claims values first
		require.Equal(t, retrievedClaims[lcom.JWTClaimAudienceKey], standardClaims.Audience)
		require.Equal(t, retrievedClaims[lcom.JWTClaimExpiresAtKey], float64(standardClaims.ExpiresAt))
		require.Equal(t, retrievedClaims[lcom.JWTClaimIDKey], standardClaims.Id)
		require.Equal(t, retrievedClaims[lcom.JWTClaimIssuedAtKey], float64(standardClaims.IssuedAt))
		require.Equal(t, retrievedClaims[lcom.JWTClaimIssuerKey], standardClaims.Issuer)
		require.Equal(t, retrievedClaims[lcom.JWTClaimNotBeforeKey], float64(standardClaims.NotBefore))
		require.Equal(t, retrievedClaims[lcom.JWTClaimSubjectKey], standardClaims.Subject)

		// verify the custom claim values second
		require.Equal(t, retrievedClaims["hi"], "sean")
		require.Equal(t, retrievedClaims["hello"], "there")
		require.Equal(t, retrievedClaims["number"], float64(34))
	})
}

func TestExtractJWT(t *testing.T) {
	standardClaims := util.GenerateStandardMapClaims()
	signedJWT, err := Sign(standardClaims)
	require.Nil(t, err)

	t.Run("verify ExtractJWT returns err for empty Authorization header", func(t *testing.T) {
		headers := map[string]string{"Authorization": ""}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, lcom.ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for Authorization header misspelled - all caps", func(t *testing.T) {
		headers := map[string]string{"AUTHORIZATION": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, lcom.ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for Authorization header misspelled - lowercase", func(t *testing.T) {
		headers := map[string]string{"authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, lcom.ErrNoAuthorizationHeader))
	})
	t.Run("verify ExtractJWT returns err for bearer prefix not used", func(t *testing.T) {
		headers := map[string]string{"Authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, lcom.ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for bearer not camel cased", func(t *testing.T) {
		headers := map[string]string{"Authorization": "bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, lcom.ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for BEARER all caps", func(t *testing.T) {
		headers := map[string]string{"Authorization": "BEARER " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, lcom.ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns err for Bearer does not end with space", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer" + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 0)
		require.Equal(t, http.StatusBadRequest, httpStatus)
		require.NotNil(t, extractErr)
		require.True(t, errors.Is(extractErr, lcom.ErrNoBearerPrefix))
	})
	t.Run("verify ExtractJWT returns claims correctly with valid input", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		require.True(t, len(mapClaims) == 7)
		require.Equal(t, http.StatusOK, httpStatus)
		require.Nil(t, extractErr)
		require.Nil(t, extractErr)

		require.Equal(t, mapClaims[lcom.JWTClaimAudienceKey], mapClaims[lcom.JWTClaimAudienceKey])
		require.Equal(t, mapClaims[lcom.JWTClaimExpiresAtKey], mapClaims[lcom.JWTClaimExpiresAtKey])
		require.Equal(t, mapClaims[lcom.JWTClaimIDKey], mapClaims[lcom.JWTClaimIDKey])
		require.Equal(t, mapClaims[lcom.JWTClaimIssuedAtKey], mapClaims[lcom.JWTClaimIssuedAtKey])
		require.Equal(t, mapClaims[lcom.JWTClaimIssuerKey], mapClaims[lcom.JWTClaimIssuerKey])
		require.Equal(t, mapClaims[lcom.JWTClaimNotBeforeKey], mapClaims[lcom.JWTClaimNotBeforeKey])
		require.Equal(t, mapClaims[lcom.JWTClaimSubjectKey], mapClaims[lcom.JWTClaimSubjectKey])
	})
}
