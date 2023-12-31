package lambda_router

import (
	"cloud.google.com/go/civil"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

var boolRegex = regexp.MustCompile(`^1|true|on|enabled$`)

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

// UnmarshalRes should generally be used only when testing as normally you return the response
// directly to the caller and won't need to Unmarshal it. However, if you are testing locally then
// it will help you extract the response body of a lambda request and marshal it to an object.
func UnmarshalRes(res events.APIGatewayProxyResponse, target interface{}) error {
	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("invalid unmarshal target, must be pointer to struct")
	}

	return json.Unmarshal([]byte(res.Body), target)
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
		return HTTPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("invalid req body: %s", err),
		}
	}

	return nil
}

import (
"fmt"
"reflect"
"strconv"
"strings"
"time"

"cloud.google.com/go/civil"
"go.mongodb.org/mongo-driver/bson/primitive"
)

func unmarshalField(
	typeField reflect.Type,
	valueField reflect.Value,
	params map[string]string,
	multiParam map[string][]string,
	param string,
) error {
	switch typeField.Kind() {
	case reflect.Ptr:
		// Handling pointer types
		elemType := typeField.Elem()
		switch {
		case elemType == reflect.TypeOf(civil.Date{}):
			if val, ok := params[param]; ok {
				parsed, err := civil.ParseDate(val)
				if err != nil {
					return err
				}
				valueField.Set(reflect.ValueOf(&parsed))
			}
		case elemType == reflect.TypeOf(time.Time{}):
			if val, ok := params[param]; ok {
				t, err := time.Parse(time.RFC3339, val)
				if err != nil {
					return err
				}
				valueField.Set(reflect.ValueOf(&t))
			}
		case elemType == reflect.TypeOf(primitive.ObjectID{}):
			hexString, ok := params[param]
			if ok && hexString != "" {
				objectID, err := primitive.ObjectIDFromHex(hexString)
				if err != nil {
					return fmt.Errorf("invalid ObjectID: %s", err)
				}
				valueField.Set(reflect.ValueOf(&objectID))
			}
			// Add additional cases for other pointer types if necessary
		}

	case reflect.Slice:
		// Handling slice types
		sliceType := typeField.Elem()
		switch {
		case sliceType == reflect.TypeOf(primitive.ObjectID{}):
			strValues, ok := multiParam[param]
			if ok {
				objectIDs := make([]primitive.ObjectID, 0, len(strValues))
				for _, str := range strValues {
					if str == "" {
						continue
					}
					objectID, err := primitive.ObjectIDFromHex(str)
					if err != nil {
						return fmt.Errorf("invalid ObjectID in slice: %s", err)
					}
					objectIDs = append(objectIDs, objectID)
				}
				valueField.Set(reflect.ValueOf(objectIDs))
			}
			// Add additional cases for other slice types if necessary
		}

	default:
		// Handling basic types
		switch typeField {
		case reflect.TypeOf(""):
			valueField.SetString(params[param])
		case reflect.TypeOf(0), reflect.TypeOf(int8(0)), reflect.TypeOf(int16(0)), reflect.TypeOf(int32(0)), reflect.TypeOf(int64(0)):
			val, err := strconv.ParseInt(params[param], 10, 64)
			if err != nil {
				return err
			}
			valueField.SetInt(val)
		case reflect.TypeOf(uint(0)), reflect.TypeOf(uint8(0)), reflect.TypeOf(uint16(0)), reflect.TypeOf(uint32(0)), reflect.TypeOf(uint64(0)):
			val, err := strconv.ParseUint(params[param], 10, 64)
			if err != nil {
				return err
			}
			valueField.SetUint(val)
		case reflect.TypeOf(float32(0)), reflect.TypeOf(float64(0)):
			val, err := strconv.ParseFloat(params[param], 64)
			if err != nil {
				return err
			}
			valueField.SetFloat(val)
		case reflect.TypeOf(false):
			val, err := strconv.ParseBool(params[param])
			if err != nil {
				return err
			}
			valueField.SetBool(val)
			// Add additional cases for other basic types if necessary
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
		return value, HTTPError{
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
		return value, HTTPError{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("%s must be a valid floating point number", param),
		}
	}

	return value, nil
}
