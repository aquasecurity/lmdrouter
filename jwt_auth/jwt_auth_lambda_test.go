package jwt_auth

import (
	"errors"
	"github.com/jgroeneveld/trial/assert"
	"net/http"
	"testing"
)

func TestExtractJWT(t *testing.T) {
	standardClaims := GenerateStandardMapClaims()
	signedJWT, err := Sign(standardClaims)
	assert.Nil(t, err)

	t.Run("verify ExtractJWT returns err for empty Authorization header", func(t *testing.T) {
		headers := map[string]string{"Authorization": ""}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})

	t.Run("verify ExtractJWT returns err for  Authorization header misspelled - all caps", func(t *testing.T) {
		headers := map[string]string{"AUTHORIZATION": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})

	t.Run("verify ExtractJWT returns err for Authorization header misspelled - lowercase", func(t *testing.T) {
		headers := map[string]string{"authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, extractErr)
		assert.True(t, errors.Is(extractErr, ErrNoAuthorizationHeader))
	})

	t.Run("verify ExtractJWT returns err for bearer prefix not used", func(t *testing.T) {
		headers := map[string]string{"Authorization": signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, extractErr)
		assert.Equal(t, extractErr.Error(), "missing 'Bearer ' prefix for Authorization header")
	})
	t.Run("verify ExtractJWT returns err for bearer not camel cased", func(t *testing.T) {
		headers := map[string]string{"Authorization": "bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, extractErr)
		assert.Equal(t, extractErr.Error(), "missing 'Bearer ' prefix for Authorization header")
	})
	t.Run("verify ExtractJWT returns err for BEARER all caps", func(t *testing.T) {
		headers := map[string]string{"Authorization": "BEARER " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, extractErr)
		assert.Equal(t, extractErr.Error(), "missing 'Bearer ' prefix for Authorization header")
	})
	t.Run("verify ExtractJWT returns err for Bearer does not end with space", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer" + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 0)
		assert.Equal(t, http.StatusUnauthorized, httpStatus)
		assert.NotNil(t, extractErr)
		assert.Equal(t, extractErr.Error(), "missing 'Bearer ' prefix for Authorization header")
	})
	t.Run("verify ExtractJWT returns claims correctly with valid input", func(t *testing.T) {
		headers := map[string]string{"Authorization": "Bearer " + signedJWT}
		mapClaims, httpStatus, extractErr := ExtractJWT(headers)
		assert.True(t, len(mapClaims) == 7)
		assert.Equal(t, http.StatusOK, httpStatus)
		assert.Nil(t, extractErr)
		assert.Nil(t, extractErr)

		assert.Equal(t, mapClaims[AudienceKey], mapClaims[AudienceKey])
		assert.Equal(t, mapClaims[ExpiresAtKey], mapClaims[ExpiresAtKey])
		assert.Equal(t, mapClaims[IDKey], mapClaims[IDKey])
		assert.Equal(t, mapClaims[IssuedAtKey], mapClaims[IssuedAtKey])
		assert.Equal(t, mapClaims[IssuerKey], mapClaims[IssuerKey])
		assert.Equal(t, mapClaims[NotBeforeKey], mapClaims[NotBeforeKey])
		assert.Equal(t, mapClaims[SubjectKey], mapClaims[SubjectKey])
	})
}
