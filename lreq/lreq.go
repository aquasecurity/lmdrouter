package lreq

import (
	"cloud.google.com/go/civil"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/seantcanavan/lambda_jwt_router/lres"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var boolRegex = regexp.MustCompile(`^1|true|on|enabled|t$`)

// UnmarshalReq "fills" out a target Go struct with data from the req.
// If body is true, then the req body is assumed to be JSON and simply
// unmarshalled into the target (taking into account that the req body may
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
func UnmarshalReq(req events.APIGatewayProxyRequest, body bool, target interface{}) error {
	if body {
		err := unmarshalBody(req, target)
		if err != nil {
			return err
		}
	}

	return unmarshalEvent(req, target)
}

func unmarshalBody(req events.APIGatewayProxyRequest, target interface{}) (err error) {
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
		return lres.HTTPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("invalid req body: %s", err),
		}
	}

	return nil
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

func unmarshalField(
	typeField reflect.Type,
	valueField reflect.Value,
	params map[string]string,
	multiParam map[string][]string,
	param string,
) error {
	strVal, ok := params[param]
	strVals, okMulti := multiParam[param]

	// check for empty / unset values and return no error if so
	if !ok && !okMulti || (strVal == "" && strVals == nil) {
		return nil
	}

	//fmt.Println(fmt.Sprintf("param %s", param))
	//fmt.Println(fmt.Sprintf("params[param] %s", strVal))
	//fmt.Println(fmt.Sprintf("multiParam[param] %+v", strVals))
	//fmt.Println(fmt.Sprintf("typeField.Name() %s", typeField.Name()))
	//fmt.Println(fmt.Sprintf("typeField.Kind() %s", typeField.Kind()))
	//
	//if typeField.Kind() == reflect.Array ||
	//	typeField.Kind() == reflect.Chan ||
	//	typeField.Kind() == reflect.Map ||
	//	typeField.Kind() == reflect.Ptr ||
	//	typeField.Kind() == reflect.Slice {
	//	fmt.Println(fmt.Sprintf("typeField.Elem() %s", typeField.Elem()))
	//	fmt.Println(fmt.Sprintf("typeField.Elem().Kind() %s", typeField.Elem().Kind()))
	//}

	//fmt.Println(fmt.Sprintf("valueField.Type() %s", valueField.Type()))
	//fmt.Println(fmt.Sprintf("valueField.Kind() %s", valueField.Kind()))
	//
	//fmt.Print("\n\n\n")

	switch typeField.Kind() {
	case reflect.Array:
		objectID, err := primitive.ObjectIDFromHex(strVal)
		if err != nil {
			return fmt.Errorf("invalid ObjectID: %s", err)
		}
		valueField.Set(reflect.ValueOf(objectID))

	case reflect.String:
		valueField.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := parseInt64Param(param, strVal, ok)
		if err != nil {
			return err
		}
		valueField.SetInt(value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := parseUint64Param(param, strVal, ok)
		if err != nil {
			return err
		}
		valueField.SetUint(value)
	case reflect.Float32, reflect.Float64:
		value, err := parseFloat64Param(param, strVal, ok)
		if err != nil {
			return err
		}
		valueField.SetFloat(value)
	case reflect.Bool:
		valueField.SetBool(boolRegex.MatchString(strings.ToLower(strVal)))
	case reflect.Ptr:
		if ok {
			switch typeField.Elem().Kind() {
			case reflect.String:
				valueField.Set(reflect.ValueOf(&strVal).Convert(typeField))
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				value, err := parseInt64Param(param, strVal, ok)
				if err != nil {
					return err
				}
				// Create a new pointer to the integer type
				intPtr := reflect.New(typeField.Elem())
				// Set the value to the newly created pointer
				intPtr.Elem().SetInt(value)
				// Set the field to the new pointer
				valueField.Set(intPtr)
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				value, err := parseUint64Param(param, strVal, ok)
				if err != nil {
					return err
				}
				// Create a new pointer to the integer type
				intPtr := reflect.New(typeField.Elem())
				// Set the value to the newly created pointer
				intPtr.Elem().SetUint(value)
				// Set the field to the new pointer
				valueField.Set(intPtr)
			case reflect.Float32, reflect.Float64:
				value, err := parseFloat64Param(param, strVal, ok)
				if err != nil {
					return err
				}
				// Create a new pointer to the integer type
				intPtr := reflect.New(typeField.Elem())
				// Set the value to the newly created pointer
				intPtr.Elem().SetFloat(value)
				// Set the field to the new pointer
				valueField.Set(intPtr)
			case reflect.Struct:
				if typeField.Elem() == reflect.TypeOf(civil.Date{}) {
					parsedCivil, err := civil.ParseDate(strVal)
					if err != nil {
						return err
					}
					valueField.Set(reflect.ValueOf(&parsedCivil))
				} else if typeField.Elem() == reflect.TypeOf(time.Time{}) {
					parsedTime, err := time.Parse(time.RFC3339, strVal)
					if err != nil {
						return err
					}
					valueField.Set(reflect.ValueOf(&parsedTime))
				}
			case reflect.Bool:
				b := boolRegex.MatchString(strings.ToLower(strVal))
				valueField.Set(reflect.ValueOf(&b))
			// Handling mongo DB ID types
			default:
				switch typeField.Elem() {
				case reflect.TypeOf(primitive.ObjectID{}):
					objectID, err := primitive.ObjectIDFromHex(strVal)
					if err != nil {
						return fmt.Errorf("invalid ObjectID: %s", err)
					}
					valueField.Set(reflect.ValueOf(&objectID))
				}
			}
		}
	case reflect.Slice:
		if typeField.Elem().Kind() == reflect.Ptr && typeField.Elem().Elem().Kind() == reflect.String {
			// Handling the slice of pointers to custom string type (like Number)
			stringValues := strVals
			if !okMulti {
				stringValues = strings.Split(strVal, ",")
			}
			slice := reflect.MakeSlice(typeField, len(stringValues), len(stringValues))

			for i, strVal := range stringValues {
				// Create a new instance of the element type (which is a pointer)
				newElemPtr := reflect.New(typeField.Elem().Elem())
				// Set the value of the new instance
				newElemPtr.Elem().SetString(strVal)
				// Set the slice element to the new instance
				slice.Index(i).Set(newElemPtr)
			}

			valueField.Set(slice)
		} else {
			// we'll be extracting values from multiParam, generating a slice and
			// putting it in valueField
			if okMulti {
				slice := reflect.MakeSlice(typeField, len(strVals), len(strVals))

				for i, str := range strVals {
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
				if ok {
					stringParts := strings.Split(strVal, ",")
					if len(stringParts) < 1 {
						return nil
					}
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

	case reflect.Struct:
		switch valueField.Type() {
		case reflect.TypeOf(time.Time{}):
			parsedTime, err := time.Parse(time.RFC3339, strVal)
			if err != nil {
				return err
			}
			valueField.Set(reflect.ValueOf(parsedTime))
		case reflect.TypeOf(civil.Date{}):
			parsedCivil, err := civil.ParseDate(strVal)
			if err != nil {
				return err
			}
			valueField.Set(reflect.ValueOf(parsedCivil))
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
		return value, lres.HTTPError{
			Status:  http.StatusBadRequest,
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
		return value, lres.HTTPError{
			Status:  http.StatusBadRequest,
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
		return value, lres.HTTPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("%s must be a valid floating point number", param),
		}
	}

	return value, nil
}

// MarshalReq will take an interface input, marshal it to JSON, and add the
// JSON as a string to the events.APIGatewayProxyRequest body field before returning.
func MarshalReq(input interface{}) events.APIGatewayProxyRequest {
	jsonBytes, err := json.Marshal(input)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	return events.APIGatewayProxyRequest{
		Body: string(jsonBytes),
	}
}
