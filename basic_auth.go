package lmdrouter

import (
	"encoding/base64"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

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
