package lmdrouter

import "fmt"

// HTTPError is a generic struct type for JSON error responses. It allows the
// library to assign an HTTP status code for the errors returned by its various
// functions.
type HTTPError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

// Error returns a string representation of the error.
func (err HTTPError) Error() string {
	return fmt.Sprintf("error %d: %s", err.Status, err.Message)
}
