package lmdrouter

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

// BasicAuth attempts to parse a username and password from the request's
// "Authorization" header. If the Authorization header is missing, or does not
// contain valid Basic HTTP Authentication date, empty values will be returned.
func BasicAuth(req events.APIGatewayProxyRequest) (user, pass string) {
	auth := req.Headers["Authorization"]
	if auth == "" {
		return user, pass
	}

	const prefix = "Basic "
	if len(auth) < len(prefix) || !strings.EqualFold(auth[:len(prefix)], prefix) {
		return user, pass
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return user, pass
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return user, pass
	}

	return cs[:s], cs[s+1:]
}
