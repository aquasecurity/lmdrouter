package lmdrouter

import (
	"github.com/jgroeneveld/trial/assert"
	"net/http"
	"testing"
	"time"
)

func TestJwtAuth(t *testing.T) {
	t.Run("no adminID", func(t *testing.T) {
		adminID := ""
		adminLevel := "Developer"
		expirationDate := time.Now().Add(time.Minute * 1)
		jwt, httpStatus, err := GenerateJwt(adminID, adminLevel, &expirationDate)
		assert.Equal(t, httpStatus, http.StatusBadRequest)
		assert.NotNil(t, err)
		assert.Equal(t, jwt, "")
	})

	t.Run("no adminLevel", func(t *testing.T) {
		adminID := GenerateRandomString(10)
		adminLevel := ""
		expirationDate := time.Now().Add(time.Minute * 1)
		jwt, httpStatus, err := GenerateJwt(adminID, adminLevel, &expirationDate)
		assert.Equal(t, httpStatus, http.StatusBadRequest)
		assert.NotNil(t, err)
		assert.Equal(t, jwt, "")
	})

	t.Run("no expirationDate", func(t *testing.T) {
		adminID := GenerateRandomString(10)
		adminLevel := "Developer"
		var expirationDate time.Time
		jwt, httpStatus, err := GenerateJwt(adminID, adminLevel, &expirationDate)
		assert.Equal(t, httpStatus, http.StatusBadRequest)
		assert.NotNil(t, err)
		assert.Equal(t, jwt, "")
	})

	t.Run("expirationDate before now (already expired)", func(t *testing.T) {
		adminID := GenerateRandomString(10)
		adminLevel := "Developer"
		expirationDate := time.Now().Add(time.Minute * -10)
		jwt, httpStatus, err := GenerateJwt(adminID, adminLevel, &expirationDate)
		assert.Equal(t, httpStatus, http.StatusBadRequest)
		assert.NotNil(t, err)
		assert.Equal(t, jwt, "")
	})

	t.Run("successfully encode JWT, make claim, and Verify JWT and compare values against claim", func(t *testing.T) {
		adminID := GenerateRandomString(10)
		adminLevel := "Developer"
		expirationDate := time.Now().Add(time.Minute * 1)
		jwt, httpStatus, err := GenerateJwt(adminID, adminLevel, &expirationDate)
		assert.Equal(t, httpStatus, http.StatusOK)
		assert.Nil(t, err)
		assert.NotEqual(t, jwt, "")

		adminClaim, httpStatus, err := Verify(jwt)
		assert.Equal(t, httpStatus, http.StatusOK)
		assert.Nil(t, err)
		assert.Equal(t, adminClaim.ID, adminID)
		assert.Equal(t, adminClaim.Level, adminLevel)
		assert.True(t, AreSameSecond(&adminClaim.ExpirationDate, &expirationDate))
	})
}
