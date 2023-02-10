package lambda_router

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
)

// ServerHTTP implements the net/http.Handler interface in order to allow
// lmdrouter applications to be used outside of AWS Lambda environments, most
// likely for local development purposes
func (l *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// convert req into an events.APIGatewayProxyRequest object
	singleValueHeaders := convertMap(r.Header)
	singleValueQuery := convertMap(
		r.URL.Query(),
	)

	corsMethods := os.Getenv("LAMBDA_JWT_ROUTER_CORS_METHODS")
	corsOrigins := os.Getenv("LAMBDA_JWT_ROUTER_CORS_ORIGIN")

	if len(corsMethods) == 0 {
		corsMethods = "*"
	}

	if len(corsOrigins) == 0 {
		corsOrigins = "*"
	}

	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", corsMethods)
	w.Header().Set("Access-Control-Allow-Origin", corsOrigins)
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(500)
		encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Failed reading req body: %s", err),
		})
		if encodeErr != nil {
			log.Printf("encodeErr [%+v]", encodeErr)
		}
		return
	}

	event := events.APIGatewayProxyRequest{
		Body:                            string(body),
		HTTPMethod:                      r.Method,
		Headers:                         singleValueHeaders,
		MultiValueHeaders:               map[string][]string(r.Header),
		MultiValueQueryStringParameters: map[string][]string(r.URL.Query()),
		Path:                            r.URL.Path,
		QueryStringParameters:           singleValueQuery,
	}

	res, err := l.Handler(r.Context(), event)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(500)
		encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
			"error": fmt.Sprintf("Failed executing handler: %s", err),
		})
		if encodeErr != nil {
			log.Printf("encodeErr [%+v]", encodeErr)
		}
		return
	}

	var resBody []byte
	if res.IsBase64Encoded {
		resBody, err = base64.StdEncoding.DecodeString(res.Body)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(500)
			encodeErr := json.NewEncoder(w).Encode(map[string]interface{}{
				"error": fmt.Sprintf("Handler returned invalid base64 data: %s", err),
			})
			if encodeErr != nil {
				log.Printf("encodeErr [%+v]", encodeErr)
			}
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
	_, writeErr := w.Write(resBody)
	if writeErr != nil {
		log.Printf("writeErr [%+v]", writeErr)
	}
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
