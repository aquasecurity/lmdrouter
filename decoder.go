package lmdrouter

// nolint: unused

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

var boolRegex = regexp.MustCompile(`^1|true|on|enabled$`)

// UnmarshalRequest "fills" out a target Go struct with data from the request.
// If body is true, then the request body is assumed to be JSON and simply
// unmarshaled into the target (taking into account that the request body may
// be base-64 encoded). If body is false, the function will traverse the
// exported fields of the target struct, and fill those that include the
// "lambda" struct tag with values taken from the request's query string
// parameters, path parameters or headers, according to the tag definition.
// Field types are currently limited to string, all int types, all uint
// types, and bool (that also means nested structs are not currently supported).
//
// Example struct:
//
//     type ListPostsInput struct {
//         ID          uint64 `lambda:"path.id"`
//         Page        uint64 `lambda:"query.page"`
//         PageSize    uint64 `lambda:"query.page_size"`
//         Search      string `lambda:"query.search"`
//         ShowDrafts  bool   `lambda:"query.show_hidden"`
//         Language    string `lambda:"header.Accept-Language"`
//     }
//
func UnmarshalRequest(
	req events.APIGatewayProxyRequest,
	body bool,
	target interface{},
) error {
	if body {
		return unmarshalBody(req, target)
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("invalid unmarshal target, must be pointer to struct")
	}

	v := rv.Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		typeField := t.Field(i)
		valueField := v.Field(i)

		lambdaTag := typeField.Tag.Get("lambda")
		if lambdaTag == "" {
			continue
		}

		components := strings.Split(lambdaTag, ".")
		if len(components) != 2 {
			return fmt.Errorf("invalid lambda tag for field %s", typeField.Name)
		}

		var sourceMap map[string]string

		switch components[0] {
		case "query":
			sourceMap = req.QueryStringParameters
		case "path":
			sourceMap = req.PathParameters
		case "header":
			sourceMap = req.Headers
		default:
			return fmt.Errorf(
				"invalid param location %q for field %s",
				components[0], typeField.Name,
			)
		}

		err := unmarshalField(typeField, valueField, sourceMap, components[1])
		if err != nil {
			return err
		}
	}
	return nil
}

func unmarshalBody(req events.APIGatewayProxyRequest, target interface{}) (
	err error,
) {
	if req.IsBase64Encoded {
		var body []byte
		body, err = base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return fmt.Errorf("failed decoding body: %w", err)
		}

		err = json.Unmarshal(body, target)
	} else {
		err = json.Unmarshal([]byte(req.Body), target)
	}

	if err != nil {
		return HTTPError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("invalid request body: %s", err),
		}
	}

	return nil
}

func unmarshalField(
	typeField reflect.StructField,
	valueField reflect.Value,
	params map[string]string,
	param string,
) error {
	switch typeField.Type.Kind() {
	case reflect.String:
		valueField.SetString(params[param])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := parseInt64Param(params, param)
		if err != nil {
			return err
		}
		valueField.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := parseUint64Param(params, param)
		if err != nil {
			return err
		}
		valueField.SetUint(value)
	case reflect.Bool:
		valueField.SetBool(boolRegex.MatchString(params[param]))
	}

	return nil
}

func parseInt64Param(params map[string]string, param string) (
	value int64,
	err error,
) {
	str, ok := params[param]
	if !ok {
		return value, nil
	}

	value, err = strconv.ParseInt(str, 10, 64)
	if err != nil {
		return value, HTTPError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("%s must be a valid integer", param),
		}
	}

	return value, nil
}

func parseUint64Param(params map[string]string, param string) (
	value uint64,
	err error,
) {
	str, ok := params[param]
	if !ok {
		return value, nil
	}

	value, err = strconv.ParseUint(str, 10, 64)
	if err != nil {
		return value, HTTPError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("%s must be a valid, positive integer", param),
		}
	}

	return value, nil
}
