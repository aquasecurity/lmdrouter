package lmdrouter

import (
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/jgroeneveld/trial/assert"
)

func TestBasicAuth(t *testing.T) {
	t.Run("no auth header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{}
		user, pass := BasicAuth(req)
		assert.Equal(t, "", user, "user must be empty")
		assert.Equal(t, "", pass, "pass must be empty")
	})

	t.Run("invalid auth header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"Authorization": "Bla",
			},
		}
		user, pass := BasicAuth(req)
		assert.Equal(t, "", user, "user must be empty")
		assert.Equal(t, "", pass, "pass must be empty")
	})

	t.Run("valid auth header", func(t *testing.T) {
		req := events.APIGatewayProxyRequest{
			Headers: map[string]string{
				"Authorization": "Basic dXNlcm5hbWU6cGFzc3BocmFzZQ==",
			},
		}
		user, pass := BasicAuth(req)
		assert.Equal(t, "username", user, "user must be correct")
		assert.Equal(t, "passphrase", pass, "pass must be correct")
	})
}
