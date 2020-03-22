// nolint: unused
package lmdrouter

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var boolRegex = regexp.MustCompile(`^1|true|on|enabled$`)

func UnmarshalRequest(
	pathParams map[string]string,
	queryParams map[string]string,
	target interface{},
) error {
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

		switch components[0] {
		case "query":
			err := unmarshalField(typeField, valueField, queryParams, components[1])
			if err != nil {
				return err
			}
		case "path":
			err := unmarshalField(typeField, valueField, pathParams, components[1])
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf(
				"invalid param location %s for field %s",
				components[0], typeField.Name,
			)
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
