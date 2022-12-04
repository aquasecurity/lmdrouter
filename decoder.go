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
	"time"

	"github.com/aws/aws-lambda-go/events"
)

var boolRegex = regexp.MustCompile(`^1|true|on|enabled$`)

// UnmarshalRequest "fills" out a target Go struct with data from the request.
// If body is true, then the request body is assumed to be JSON and simply
// unmarshaled into the target (taking into account that the request body may
// be base-64 encoded). After that, or if body is false, the function will
// traverse the exported fields of the target struct, and fill those that
// include the "lambda" struct tag with values taken from the request's query
// string parameters, path parameters and headers, according to the field's
// struct tag definition. This means a struct value can be filled with data from
// the body, the path, the query string and the headers at the same time.
//
// Field types are currently limited to string, all integer types, all unsigned
// integer types, all float types, booleans, slices of the aforementioned types
// and pointers of these types.
//
// Note that custom types that alias any of the aforementioned types are also
// accepted and the appropriate constant values will be generated. Boolean
// fields accept (in a case-insensitive way) the values "1", "true", "on" and
// "enabled". Any other value is considered false.
//
// Example struct (no body):
//
//	type ListPostsInput struct {
//	    ID          uint64   `lambda:"path.id"`
//	    Page        uint64   `lambda:"query.page"`
//	    PageSize    uint64   `lambda:"query.page_size"`
//	    Search      string   `lambda:"query.search"`
//	    ShowDrafts  bool     `lambda:"query.show_hidden"`
//	    Languages   []string `lambda:"header.Accept-Language"`
//	}
//
// Example struct (JSON body):
//
//	type UpdatePostInput struct {
//	    ID          uint64   `lambda:"path.id"`
//	    Author      string   `lambda:"header.Author"`
//	    Title       string   `json:"title"`
//	    Content     string   `json:"content"`
//	}
func UnmarshalRequest(
	req events.APIGatewayProxyRequest,
	body bool,
	target interface{},
) error {
	if body {
		err := unmarshalBody(req, target)
		if err != nil {
			return err
		}
	}

	return unmarshalEvent(req, target)
}

func unmarshalEvent(req events.APIGatewayProxyRequest, target interface{}) error {
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
		var multiMap map[string][]string

		switch components[0] {
		case "query":
			sourceMap = req.QueryStringParameters
			multiMap = req.MultiValueQueryStringParameters
		case "path":
			sourceMap = req.PathParameters
		case "header":
			sourceMap = req.Headers
			multiMap = req.MultiValueHeaders
		default:
			return fmt.Errorf(
				"invalid param location %q for field %s",
				components[0], typeField.Name,
			)
		}

		err := unmarshalField(
			typeField.Type,
			valueField,
			sourceMap,
			multiMap,
			components[1],
		)
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
	typeField reflect.Type,
	valueField reflect.Value,
	params map[string]string,
	multiParam map[string][]string,
	param string,
) error {
	switch typeField.Kind() {
	case reflect.String:
		valueField.SetString(params[param])
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		str, ok := params[param]
		value, err := parseInt64Param(param, str, ok)
		if err != nil {
			return err
		}
		valueField.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		str, ok := params[param]
		value, err := parseUint64Param(param, str, ok)
		if err != nil {
			return err
		}
		valueField.SetUint(value)
	case reflect.Float32, reflect.Float64:
		str, ok := params[param]
		value, err := parseFloat64Param(param, str, ok)
		if err != nil {
			return err
		}
		valueField.SetFloat(value)
	case reflect.Bool:
		valueField.SetBool(boolRegex.MatchString(strings.ToLower(params[param])))
	case reflect.Ptr:
		if val, ok := params[param]; ok {
			switch typeField.Elem().Kind() {
			case reflect.Int:
				value, err := parseInt64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := int(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Int8:
				value, err := parseInt64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := int8(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Int16:
				value, err := parseInt64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := int16(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Int32:
				value, err := parseInt64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := int32(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Int64:
				value, err := parseInt64Param(param, val, ok)
				if err != nil {
					return err
				}
				valueField.Set(reflect.ValueOf(&value))
			case reflect.Uint:
				value, err := parseUint64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := uint(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Uint8:
				value, err := parseUint64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := uint8(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Uint16:
				value, err := parseUint64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := uint16(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Uint32:
				value, err := parseUint64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := uint32(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Uint64:
				value, err := parseUint64Param(param, val, ok)
				if err != nil {
					return err
				}
				valueField.Set(reflect.ValueOf(&value))
			case reflect.Float32:
				value, err := parseFloat64Param(param, val, ok)
				if err != nil {
					return err
				}
				convertedValue := float32(value)
				valueField.Set(reflect.ValueOf(&convertedValue))
			case reflect.Float64:
				value, err := parseFloat64Param(param, val, ok)
				if err != nil {
					return err
				}
				valueField.Set(reflect.ValueOf(&value))
			case reflect.String:
				valueField.Set(reflect.ValueOf(&val).Convert(typeField))
			case reflect.Struct:
				if typeField.Elem() == reflect.TypeOf(time.Now()) {
					parsedTime, err := time.Parse(time.RFC3339, val)
					if err != nil {
						return err
					}
					valueField.Set(reflect.ValueOf(&parsedTime))
				}
			case reflect.Bool:
				b := boolRegex.MatchString(strings.ToLower(val))
				valueField.Set(reflect.ValueOf(&b))
			}
		}
	case reflect.Slice:
		// we'll be extracting values from multiParam, generating a slice and
		// putting it in valueField
		strs, ok := multiParam[param]
		if ok {
			slice := reflect.MakeSlice(typeField, len(strs), len(strs))

			for i, str := range strs {
				err := unmarshalField(
					typeField.Elem(),
					slice.Index(i),
					map[string]string{"param": str},
					nil,
					"param",
				)
				if err != nil {
					return err
				}
			}

			valueField.Set(slice)
		} else {
			str, ok := params[param]
			if ok {
				stringParts := strings.Split(str, ",")
				slice := reflect.MakeSlice(typeField, len(stringParts), len(stringParts))

				for i, p := range stringParts {
					inVal := reflect.ValueOf(p)
					asVal := inVal.Convert(typeField.Elem())
					slice.Index(i).Set(asVal)
				}

				valueField.Set(slice)
			}
		}
	}

	return nil
}

func parseInt64Param(param, str string, ok bool) (value int64, err error) {
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

func parseUint64Param(param, str string, ok bool) (value uint64, err error) {
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

func parseFloat64Param(param, str string, ok bool) (value float64, err error) {
	if !ok {
		return value, nil
	}

	value, err = strconv.ParseFloat(str, 64)
	if err != nil {
		return value, HTTPError{
			Code:    http.StatusBadRequest,
			Message: fmt.Sprintf("%s must be a valid floating point number", param),
		}
	}

	return value, nil
}
