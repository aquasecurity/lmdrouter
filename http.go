package lmdrouter

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
)

// ServerHTTP implements the net/http.Handler interface in order to allow
// lmdrouter applications to be used outside of AWS Lambda environments, most
// likely for local development purposes
func (l *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// convert request into an events.APIGatewayProxyRequest object
	singleValueHeaders := convertMap(map[string][]string(r.Header))
	singleValueQuery := convertMap(
		map[string][]string(r.URL.Query()),
	)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Failed reading request body: %s", err),
		}) // nolint: errcheck
		return
	}

	event := events.APIGatewayProxyRequest{
		Path:                            r.URL.Path,
		HTTPMethod:                      r.Method,
		Headers:                         singleValueHeaders,
		MultiValueHeaders:               map[string][]string(r.Header),
		QueryStringParameters:           singleValueQuery,
		MultiValueQueryStringParameters: map[string][]string(r.URL.Query()),
		Body:                            string(body),
	}

	res, err := l.Handler(r.Context(), event)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Failed executing handler: %s", err),
		}) // nolint: errcheck
		return
	}

	var resBody []byte
	if res.IsBase64Encoded {
		resBody, err = base64.StdEncoding.DecodeString(res.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Handler returned invalid base64 data: %s", err),
			}) // nolint: errcheck
			return
		}
	} else {
		resBody = []byte(res.Body)
	}

	for header, values := range res.MultiValueHeaders {
		for i, value := range values {
			if i == 0 {
				w.Header().Set(header, value)
			} else {
				w.Header().Add(header, value)
			}
		}
	}

	for header, value := range res.Headers {
		if w.Header().Get(header) == "" {
			w.Header().Set(header, value)
		}
	}

	w.WriteHeader(res.StatusCode)
	w.Write(resBody) // nolint: errcheck
}

func convertMap(in map[string][]string) map[string]string {
	singleValue := make(map[string]string)

	for key, value := range in {
		if len(value) == 1 {
			singleValue[key] = value[0]
		}
	}

	return singleValue
}
