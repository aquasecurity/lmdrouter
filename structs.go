package lmdrouter

import "fmt"

type HTTPError struct {
	Code    int
	Message string
}

func (err HTTPError) Error() string {
	return fmt.Sprintf("error %d: %s", err.Code, err.Message)
}
