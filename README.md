# Lambda JWT Router
Simple HTTP router for working with Json Web Tokens (JWTs) on AWS Lambda through APIGateway proxy requests.
Supports local HTTP routing through net/http for local testing, debugging, and development.
Supports setting an HMAC secret for signing and verifying JWTs.
Supports automatically replying to HTTP OPTIONS requests so calls from browsers succeed (CORS).

## Features
1. Execute unproxied Lambda HTTP requests while running in the AWS Lambda environment
2. Execute proxied Lambda HTTP requests while running in the AWS Lambda environment
3. Rapidly iterate with local development via HTTP to prevent having to deploy to AWS Lambda for testing
   1. Optional hot reloading supported via [Fresh](https://github.com/seantcanavan/fresh)
4. Implement a smarter unmarshaler that checks body, headers, path params, query params, and multi value query params to reduce lambda boilerplate via custom tags
   1. `lambda:"path.id"` - parse path parameter for value 'id'
   2. `lambda:"query.id"` - parse query string parameters for value 'id'
   3. `lambda:"header.Authorization"` - parse headers for value 'Authorization'
   4. `json:"title"` - use standard JSON tags in conjunction with lambda tags for maximum value
5. Add a set of standard responses for error and success cases to reduce lambda boilerplate
   1. `SuccessRes(interface{})` - return any standard struct as a valid lambda success response
   2. `ErrorRes(int statusCode)` - quick return with an empty response and status code
6. Add a set of custom responses for error and success cases to reduce lambda boilerplate
   1. `CustomRes(httpStatus int, headers map[string]string, data interface{}) // modify the lambda res as much as necessary for specific cases where the defaults are not correct`
7. Implement a robust set of middlewares for authentication/authorization, logging, lambda context, and more
   1. `InjectLambdaContextMW` - add standard lambda request values to the root request context
   2. `LogRequestMW` - primitive logger that will print all outgoing responses and their status code
8. Add optional support for JWT Decoding `req.Headers["Authorization"]` via HMAC secret
   1. `DecodeStandard` - decode the standard JWT claims and add them to the request's root context`
   2. `DecodeExpanded` - parse an expanded set of JWT claims such as userId and userType and add them to the request's root context`
   3. `type Handler func(context.Context, events.APIGatewayProxyRequest)` - implement your own loggers, middlewares, and JWT decoders
9. Add optional support for CORS via environment variables

## Previous README
Go HTTP router library for AWS API Gateway-invoked Lambda Functions
Forked from [aquasecurity/lmdrouter](https://github.com/aquasecurity/lmdrouter)

## Installation
1. `go get github.com/seantcanavan/lambda_jwt_router@latest`


## How to Build locally
1. `make build`

## How to Test locally
1. `make test`

## How to Use
1. set the environment variable `LAMBDA_JWT_ROUTER_NO_CORS` to `true` to disable adding a CORS OPTIONS handler to every route automatically
   1. If you do not set it manually - the default value will be `false` (all endpoints have CORS added by default)
2. set the environment variable `LAMBDA_JWT_ROUTER_CORS_METHODS` to configure which CORS methods you would like to support
   1. If you do not set it manually - the default value will be `*`
3. set the environment variable `LAMBDA_JWT_ROUTER_CORS_ORIGIN` to configure which CORS origins you would like to support
   1. If you do not set it manually - the default value will be `*`
4. set the environment variable `LAMBDA_JWT_ROUTER_CORS_HEADERS` to configure which CORS headers you would like to support
   1. If you do not set it manually - the default value will be `*`
5. set the environment variable `LAMBDA_JWT_ROUTER_HMAC_SECRET` to configure the HMAC secret used to encode/decode JWTs
6. See https://github.com/aquasecurity/lmdrouter for the original README and details

## Sample routing example - see `routing_example.go` for more detail
```
var router *lambda_router.Router

func init() {
	router = lambda_router.NewRouter("/api")

	router.Route("DELETE", "/books/:id", books.DeleteLambda)
	router.Route("GET", "/books/:id", books.GetLambda)
	router.Route("POST", "/books", books.CreateLambda)
	router.Route("PUT", "/books/:id", books.UpdateLambda)
}

func main() {
    // if we're running this in staging or production, we want to use the lambda handler on startup
	environment := os.Getenv("STAGE")
	if environment == "staging" || environment == "production" {
		lambda.Start(router.Handler)
	} else { // else, we want to start an HTTP server to listen for local development
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Ready to listen and serve on port %s", port)
		err := http.ListenAndServe(":"+port, http.HandlerFunc(router.ServeHTTP))
		if err != nil {
			panic(fmt.Sprintf("http.ListAndServe error %s", err))
		}
	}
}
```

## Sample JWT example - see `jwt_example.go` for more detail
```
var router *lambda_router.Router

func init() {
	// implement your own base middleware functions and add to the NewRouter declaration to apply to every route
	router = lambda_router.NewRouter("/api", lambda_jwt.InjectLambdaContextMW)

	// to configure middleware at the route level, add them singularly to each route
	// DecodeStandard will automagically check events.Headers["Authorization"] for a valid JWT.
	// It will look for the LAMBDA_JWT_ROUTER_HMAC_SECRET environment variable and use that to decode
	// the JWT. If decoding succeeds, it will inject all the standard claims into the context object
	// before returning so other callers can access those fields at run time.
	router.Route("DELETE", "/books/:id", books.DeleteLambda, lambda_jwt.DecodeStandard)
	router.Route("GET", "/books/:id", books.GetLambda, lambda_jwt.DecodeStandard)
	router.Route("POST", "/books", books.CreateLambda, lambda_jwt.DecodeStandard)
	router.Route("PUT", "/books/:id", books.UpdateLambda, lambda_jwt.DecodeStandard)
}

func main() {
	// if we're running this in staging or production, we want to use the lambda handler on startup
	environment := os.Getenv("STAGE")
	if environment == "staging" || environment == "production" {
		lambda.Start(router.Handler)
	} else { // else, we want to start an HTTP server to listen for local development
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Ready to listen and serve on port %s", port)
		err := http.ListenAndServe(":"+port, http.HandlerFunc(router.ServeHTTP))
		if err != nil {
			panic(fmt.Sprintf("http.ListAndServe error %s", err))
		}
	}
}
```

## Sample middleware example - see `middleware_example.go` for more detail
```
var router *lambda_router.Router

func init() {
	// implement your own base middleware functions and add to the NewRouter declaration to apply to every route
	router = lambda_router.NewRouter("/api", lambda_jwt.InjectLambdaContextMW)

	// to configure middleware at the route level, add them singularly to each route
	router.Route("DELETE", "/books/:id", books.DeleteLambda, lambda_jwt.LogRequestMW)
	router.Route("GET", "/books/:id", books.GetLambda, lambda_jwt.LogRequestMW)
	router.Route("POST", "/books", books.CreateLambda, lambda_jwt.LogRequestMW)
	router.Route("PUT", "/books/:id", books.UpdateLambda, lambda_jwt.LogRequestMW)
}

func main() {
	// if we're running this in staging or production, we want to use the lambda handler on startup
	environment := os.Getenv("STAGE")
	if environment == "staging" || environment == "production" {
		lambda.Start(router.Handler)
	} else { // else, we want to start an HTTP server to listen for local development
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		log.Printf("Ready to listen and serve on port %s", port)
		err := http.ListenAndServe(":"+port, http.HandlerFunc(router.ServeHTTP))
		if err != nil {
			panic(fmt.Sprintf("http.ListAndServe error %s", err))
		}
	}
}
```

## All tests are passing
```
go test -v ./... -count=1
?   	github.com/seantcanavan/lambda_jwt_router/internal/examples/database	[no test files]
?   	github.com/seantcanavan/lambda_jwt_router/internal/examples/jwt	[no test files]
?   	github.com/seantcanavan/lambda_jwt_router/internal/examples/middleware	[no test files]
?   	github.com/seantcanavan/lambda_jwt_router/internal/examples/routing	[no test files]
?   	github.com/seantcanavan/lambda_jwt_router/internal/util	[no test files]
=== RUN   TestBooksLambda
=== RUN   TestBooksLambda/verify_CreateLambda_is_working_as_expected
=== RUN   TestBooksLambda/verify_CreateLambda_is_working_as_expected/verify_GetLambda_is_working_as_expected
=== RUN   TestBooksLambda/verify_CreateLambda_is_working_as_expected/verify_UpdateLambda_is_working_as_expected
=== RUN   TestBooksLambda/verify_CreateLambda_is_working_as_expected/verify_UpdateLambda_is_working_as_expected/verify_DeleteLambda_is_working_as_expected
--- PASS: TestBooksLambda (0.13s)
--- PASS: TestBooksLambda/verify_CreateLambda_is_working_as_expected (0.13s)
--- PASS: TestBooksLambda/verify_CreateLambda_is_working_as_expected/verify_GetLambda_is_working_as_expected (0.03s)
--- PASS: TestBooksLambda/verify_CreateLambda_is_working_as_expected/verify_UpdateLambda_is_working_as_expected (0.07s)
--- PASS: TestBooksLambda/verify_CreateLambda_is_working_as_expected/verify_UpdateLambda_is_working_as_expected/verify_DeleteLambda_is_working_as_expected (0.03s)
=== RUN   TestBooks
=== RUN   TestBooks/verify_Create_is_working_as_expected
=== RUN   TestBooks/verify_Create_is_working_as_expected/verify_Get_is_working_as_expected
=== RUN   TestBooks/verify_Create_is_working_as_expected/verify_Update_is_working_as_expected
=== RUN   TestBooks/verify_Create_is_working_as_expected/verify_Update_is_working_as_expected/verify_Delete_is_working_as_expected
--- PASS: TestBooks (0.14s)
--- PASS: TestBooks/verify_Create_is_working_as_expected (0.14s)
--- PASS: TestBooks/verify_Create_is_working_as_expected/verify_Get_is_working_as_expected (0.03s)
--- PASS: TestBooks/verify_Create_is_working_as_expected/verify_Update_is_working_as_expected (0.08s)
--- PASS: TestBooks/verify_Create_is_working_as_expected/verify_Update_is_working_as_expected/verify_Delete_is_working_as_expected (0.04s)
PASS
ok  	github.com/seantcanavan/lambda_jwt_router/internal/examples/books	0.728s
=== RUN   TestAllowOptionsMW
=== RUN   TestAllowOptionsMW/verify_empty_OPTIONS_req_succeeds
=== RUN   TestAllowOptionsMW/verify_OPTIONS_req_succeeds_with_invalid_JWT_for_AllowOptions
=== RUN   TestAllowOptionsMW/verify_OPTIONS_req_succeeds_with_no_Authorization_header_for_AllowOptions
--- PASS: TestAllowOptionsMW (0.00s)
--- PASS: TestAllowOptionsMW/verify_empty_OPTIONS_req_succeeds (0.00s)
--- PASS: TestAllowOptionsMW/verify_OPTIONS_req_succeeds_with_invalid_JWT_for_AllowOptions (0.00s)
--- PASS: TestAllowOptionsMW/verify_OPTIONS_req_succeeds_with_no_Authorization_header_for_AllowOptions (0.00s)
=== RUN   TestDecodeAndInjectExpandedClaims
=== RUN   TestDecodeAndInjectExpandedClaims/verify_error_is_returned_by_DecodeExpanded_when_missing_Authorization_header
=== RUN   TestDecodeAndInjectExpandedClaims/verify_context_is_returned_by_DecodeExpanded_with_a_signed_JWT
--- PASS: TestDecodeAndInjectExpandedClaims (0.00s)
--- PASS: TestDecodeAndInjectExpandedClaims/verify_error_is_returned_by_DecodeExpanded_when_missing_Authorization_header (0.00s)
--- PASS: TestDecodeAndInjectExpandedClaims/verify_context_is_returned_by_DecodeExpanded_with_a_signed_JWT (0.00s)
=== RUN   TestDecodeAndInjectStandardClaims
=== RUN   TestDecodeAndInjectStandardClaims/verify_error_is_returned_by_DecodeStandard_when_missing_Authorization_header
=== RUN   TestDecodeAndInjectStandardClaims/verify_context_is_returned_by_DecodeStandard_with_a_signed_JWT
--- PASS: TestDecodeAndInjectStandardClaims (0.00s)
--- PASS: TestDecodeAndInjectStandardClaims/verify_error_is_returned_by_DecodeStandard_when_missing_Authorization_header (0.00s)
--- PASS: TestDecodeAndInjectStandardClaims/verify_context_is_returned_by_DecodeStandard_with_a_signed_JWT (0.00s)
=== RUN   TestExtractJWT
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_err_for_empty_Authorization_header
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_err_for_Authorization_header_misspelled_-_all_caps
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_err_for_Authorization_header_misspelled_-_lowercase
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_err_for_bearer_prefix_not_used
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_err_for_bearer_not_camel_cased
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_err_for_BEARER_all_caps
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_err_for_Bearer_does_not_end_with_space
=== RUN   TestExtractJWT/verify_ExtractJWT_returns_claims_correctly_with_valid_input
--- PASS: TestExtractJWT (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_err_for_empty_Authorization_header (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_err_for_Authorization_header_misspelled_-_all_caps (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_err_for_Authorization_header_misspelled_-_lowercase (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_err_for_bearer_prefix_not_used (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_err_for_bearer_not_camel_cased (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_err_for_BEARER_all_caps (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_err_for_Bearer_does_not_end_with_space (0.00s)
--- PASS: TestExtractJWT/verify_ExtractJWT_returns_claims_correctly_with_valid_input (0.00s)
=== RUN   TestGenerateEmptyErrorHandler
=== RUN   TestGenerateEmptyErrorHandler/verify_empty_error_handler_returns_error
--- PASS: TestGenerateEmptyErrorHandler (0.00s)
--- PASS: TestGenerateEmptyErrorHandler/verify_empty_error_handler_returns_error (0.00s)
=== RUN   TestGenerateEmptySuccessHandler
=== RUN   TestGenerateEmptySuccessHandler/verify_empty_success_handler_returns_success
--- PASS: TestGenerateEmptySuccessHandler (0.00s)
--- PASS: TestGenerateEmptySuccessHandler/verify_empty_success_handler_returns_success (0.00s)
=== RUN   TestExtendExpandedClaims
=== RUN   TestExtendExpandedClaims/verify_sign_and_verify_expanded_and_custom_fields_in_claims
--- PASS: TestExtendExpandedClaims (0.00s)
--- PASS: TestExtendExpandedClaims/verify_sign_and_verify_expanded_and_custom_fields_in_claims (0.00s)
=== RUN   TestExtendStandardClaims
=== RUN   TestExtendStandardClaims/verify_sign_and_verify_standard_and_custom_fields_in_claims
--- PASS: TestExtendStandardClaims (0.00s)
--- PASS: TestExtendStandardClaims/verify_sign_and_verify_standard_and_custom_fields_in_claims (0.00s)
=== RUN   TestExtractCustomClaims
=== RUN   TestExtractCustomClaims/verify_ExtractCustom_returns_an_err_when_unmarshalling_to_invalid_custom_claims_object
=== RUN   TestExtractCustomClaims/verify_ExtractCustom_works_when_called_with_the_correct_parameters
--- PASS: TestExtractCustomClaims (0.00s)
--- PASS: TestExtractCustomClaims/verify_ExtractCustom_returns_an_err_when_unmarshalling_to_invalid_custom_claims_object (0.00s)
--- PASS: TestExtractCustomClaims/verify_ExtractCustom_works_when_called_with_the_correct_parameters (0.00s)
=== RUN   TestExtractStandardClaims
=== RUN   TestExtractStandardClaims/verify_ExtractStandard_returns_an_err_when_unmarshalling_to_invalid_standard_claims_object
=== RUN   TestExtractStandardClaims/verify_ExtractCustom_works_when_called_with_the_correct_parameters
--- PASS: TestExtractStandardClaims (0.00s)
--- PASS: TestExtractStandardClaims/verify_ExtractStandard_returns_an_err_when_unmarshalling_to_invalid_standard_claims_object (0.00s)
--- PASS: TestExtractStandardClaims/verify_ExtractCustom_works_when_called_with_the_correct_parameters (0.00s)
=== RUN   TestSign
=== RUN   TestSign/verify_signed_jwt_secret_with_valid_standard_claim
--- PASS: TestSign (0.00s)
--- PASS: TestSign/verify_signed_jwt_secret_with_valid_standard_claim (0.00s)
=== RUN   TestVerifyJWT
=== RUN   TestVerifyJWT/verify_err_when_parsing_invalid_jwt
=== RUN   TestVerifyJWT/verify_err_when_parsing_expired_token_with_valid_jwt
--- PASS: TestVerifyJWT (0.00s)
--- PASS: TestVerifyJWT/verify_err_when_parsing_invalid_jwt (0.00s)
--- PASS: TestVerifyJWT/verify_err_when_parsing_expired_token_with_valid_jwt (0.00s)
PASS
ok  	github.com/seantcanavan/lambda_jwt_router/lambda_jwt	0.004s
=== RUN   TestMarshalLambdaRequest
=== RUN   TestMarshalLambdaRequest/verify_MarshalReq_correctly_adds_the_JSON_string_to_the_request_body
--- PASS: TestMarshalLambdaRequest (0.00s)
--- PASS: TestMarshalLambdaRequest/verify_MarshalReq_correctly_adds_the_JSON_string_to_the_request_body (0.00s)
=== RUN   Test_UnmarshalReq
=== RUN   Test_UnmarshalReq/valid_path&query_input
=== RUN   Test_UnmarshalReq/valid_empty_input
=== RUN   Test_UnmarshalReq/valid_input_unset_values
=== RUN   Test_UnmarshalReq/invalid_path&query_input
=== RUN   Test_UnmarshalReq/valid_body_input,_not_base64
=== RUN   Test_UnmarshalReq/invalid_body_input,_not_base64
=== RUN   Test_UnmarshalReq/valid_body_input,_base64
=== RUN   Test_UnmarshalReq/invalid_body_input,_base64
--- PASS: Test_UnmarshalReq (0.00s)
--- PASS: Test_UnmarshalReq/valid_path&query_input (0.00s)
--- PASS: Test_UnmarshalReq/valid_empty_input (0.00s)
--- PASS: Test_UnmarshalReq/valid_input_unset_values (0.00s)
--- PASS: Test_UnmarshalReq/invalid_path&query_input (0.00s)
--- PASS: Test_UnmarshalReq/valid_body_input,_not_base64 (0.00s)
--- PASS: Test_UnmarshalReq/invalid_body_input,_not_base64 (0.00s)
--- PASS: Test_UnmarshalReq/valid_body_input,_base64 (0.00s)
--- PASS: Test_UnmarshalReq/invalid_body_input,_base64 (0.00s)
=== RUN   TestHTTPHandler
=== RUN   TestHTTPHandler/POST_/api_without_auth
=== RUN   TestHTTPHandler/POST_/api_with_auth
=== RUN   TestHTTPHandler/GET_/api
=== RUN   TestHTTPHandler/GET_/api/something/stuff
--- PASS: TestHTTPHandler (0.00s)
--- PASS: TestHTTPHandler/POST_/api_without_auth (0.00s)
--- PASS: TestHTTPHandler/POST_/api_with_auth (0.00s)
--- PASS: TestHTTPHandler/GET_/api (0.00s)
--- PASS: TestHTTPHandler/GET_/api/something/stuff (0.00s)
=== RUN   TestCustomRes
=== RUN   TestCustomRes/verify_CustomRes_returns_the_struct_in_the_response_body
=== RUN   TestCustomRes/verify_CustomRes_returns_the_key_value_pair_in_the_response_headers
=== RUN   TestCustomRes/verify_CustomRes_returns_the_correct_status_code
=== RUN   TestCustomRes/verify_CustomRes_returns_CORS_headers
--- PASS: TestCustomRes (0.00s)
--- PASS: TestCustomRes/verify_CustomRes_returns_the_struct_in_the_response_body (0.00s)
--- PASS: TestCustomRes/verify_CustomRes_returns_the_key_value_pair_in_the_response_headers (0.00s)
--- PASS: TestCustomRes/verify_CustomRes_returns_the_correct_status_code (0.00s)
--- PASS: TestCustomRes/verify_CustomRes_returns_CORS_headers (0.00s)
=== RUN   TestEmptyRes
=== RUN   TestEmptyRes/verify_EmptyRes_returns_the_correct_status_code
=== RUN   TestEmptyRes/verify_EmptyRes_returns_CORS_headers
--- PASS: TestEmptyRes (0.00s)
--- PASS: TestEmptyRes/verify_EmptyRes_returns_the_correct_status_code (0.00s)
--- PASS: TestEmptyRes/verify_EmptyRes_returns_CORS_headers (0.00s)
=== RUN   TestErrorRes
=== RUN   TestErrorRes/Handle_an_HTTPError_ErrorRes_without_ExposeServerErrors_set_and_verify_CORS
=== RUN   TestErrorRes/Handle_an_HTTPError_ErrorRes_without_ExposeServerErrors_set_and_verify_CORS/verify_ErrorRes_returns_CORS_headers
=== RUN   TestErrorRes/Handle_an_HTTPError_for_ErrorRes_when_ExposeServerErrors_is_true
=== RUN   TestErrorRes/Handle_an_HTTPError_for_ErrorRes_when_ExposeServerErrors_is_false
=== RUN   TestErrorRes/Handle_a_general_error_for_ErrorRes_when_ExposeServerErrors_is_true
=== RUN   TestErrorRes/Handle_a_general_error_for_ErrorRes_when_ExposeServerErrors_is_false
--- PASS: TestErrorRes (0.00s)
--- PASS: TestErrorRes/Handle_an_HTTPError_ErrorRes_without_ExposeServerErrors_set_and_verify_CORS (0.00s)
--- PASS: TestErrorRes/Handle_an_HTTPError_ErrorRes_without_ExposeServerErrors_set_and_verify_CORS/verify_ErrorRes_returns_CORS_headers (0.00s)
--- PASS: TestErrorRes/Handle_an_HTTPError_for_ErrorRes_when_ExposeServerErrors_is_true (0.00s)
--- PASS: TestErrorRes/Handle_an_HTTPError_for_ErrorRes_when_ExposeServerErrors_is_false (0.00s)
--- PASS: TestErrorRes/Handle_a_general_error_for_ErrorRes_when_ExposeServerErrors_is_true (0.00s)
--- PASS: TestErrorRes/Handle_a_general_error_for_ErrorRes_when_ExposeServerErrors_is_false (0.00s)
=== RUN   TestFileRes
=== RUN   TestFileRes/verify_FileRes_returns_the_correct_status_code
=== RUN   TestFileRes/verify_FileRes_marks_the_response_as_NOT_base64_encoded
=== RUN   TestFileRes/verify_FileRes_embeds_the_bytes_correctly_in_the_response_object_as_a_string
=== RUN   TestFileRes/verify_FileRes_preserves_the_original_header_values
=== RUN   TestFileRes/verify_FileRes_returns_CORS_headers
--- PASS: TestFileRes (0.00s)
--- PASS: TestFileRes/verify_FileRes_returns_the_correct_status_code (0.00s)
--- PASS: TestFileRes/verify_FileRes_marks_the_response_as_NOT_base64_encoded (0.00s)
--- PASS: TestFileRes/verify_FileRes_embeds_the_bytes_correctly_in_the_response_object_as_a_string (0.00s)
--- PASS: TestFileRes/verify_FileRes_preserves_the_original_header_values (0.00s)
--- PASS: TestFileRes/verify_FileRes_returns_CORS_headers (0.00s)
=== RUN   TestFileB64Res
=== RUN   TestFileB64Res/verify_FileB64Res_returns_the_correct_status_code
=== RUN   TestFileB64Res/verify_FileB64Res_marks_the_response_as_base64_encoded
=== RUN   TestFileB64Res/verify_FileB64Res_embeds_the_bytes_correctly_in_the_response_object_as_a_byte64_encoded_string
=== RUN   TestFileB64Res/verify_FileRes_preserves_the_original_header_values
=== RUN   TestFileB64Res/verify_FileB64Res_returns_CORS_headers
--- PASS: TestFileB64Res (0.00s)
--- PASS: TestFileB64Res/verify_FileB64Res_returns_the_correct_status_code (0.00s)
--- PASS: TestFileB64Res/verify_FileB64Res_marks_the_response_as_base64_encoded (0.00s)
--- PASS: TestFileB64Res/verify_FileB64Res_embeds_the_bytes_correctly_in_the_response_object_as_a_byte64_encoded_string (0.00s)
--- PASS: TestFileB64Res/verify_FileRes_preserves_the_original_header_values (0.00s)
--- PASS: TestFileB64Res/verify_FileB64Res_returns_CORS_headers (0.00s)
=== RUN   TestStatusAndErrorRes
=== RUN   TestStatusAndErrorRes/verify_StatusAndErrorRes_returns_the_correct_status_code
=== RUN   TestStatusAndErrorRes/verify_StatusAndErrorRes_returns_CORS_headers
--- PASS: TestStatusAndErrorRes (0.00s)
--- PASS: TestStatusAndErrorRes/verify_StatusAndErrorRes_returns_the_correct_status_code (0.00s)
--- PASS: TestStatusAndErrorRes/verify_StatusAndErrorRes_returns_CORS_headers (0.00s)
=== RUN   TestSuccessRes
=== RUN   TestSuccessRes/verify_SuccessRes_returns_the_correct_status_code
=== RUN   TestSuccessRes/verify_SuccessRes_returns_the_struct_in_the_response_body
=== RUN   TestSuccessRes/verify_SuccessRes_returns_CORS_headers
--- PASS: TestSuccessRes (0.00s)
--- PASS: TestSuccessRes/verify_SuccessRes_returns_the_correct_status_code (0.00s)
--- PASS: TestSuccessRes/verify_SuccessRes_returns_the_struct_in_the_response_body (0.00s)
--- PASS: TestSuccessRes/verify_SuccessRes_returns_CORS_headers (0.00s)
=== RUN   TestRouter
=== RUN   TestRouter/Routes_created_correctly
=== RUN   TestRouter/Routes_created_correctly//
=== RUN   TestRouter/Routes_created_correctly//:id
=== RUN   TestRouter/Routes_created_correctly//:id/stuff/:fake
=== RUN   TestRouter/Reqs_matched_correctly
=== RUN   TestRouter/Reqs_matched_correctly/POST_/api
=== RUN   TestRouter/Reqs_matched_correctly/POST_/api/
=== RUN   TestRouter/Reqs_matched_correctly/DELETE_/api
=== RUN   TestRouter/Reqs_matched_correctly/GET_/api/fake-id
=== RUN   TestRouter/Reqs_matched_correctly/GET_/api/fake-id/bla
=== RUN   TestRouter/Reqs_matched_correctly/GET_/api/fake-id/stuff/faked-fake
=== RUN   TestRouter/Reqs_execute_correctly
=== RUN   TestRouter/Reqs_execute_correctly/POST_/api_without_auth
=== RUN   TestRouter/Reqs_execute_correctly/POST_/api_with_auth
=== RUN   TestRouter/Reqs_execute_correctly/GET_/api
=== RUN   TestRouter/Overlapping_routes
--- PASS: TestRouter (0.00s)
--- PASS: TestRouter/Routes_created_correctly (0.00s)
--- PASS: TestRouter/Routes_created_correctly// (0.00s)
--- PASS: TestRouter/Routes_created_correctly//:id (0.00s)
--- PASS: TestRouter/Routes_created_correctly//:id/stuff/:fake (0.00s)
--- PASS: TestRouter/Reqs_matched_correctly (0.00s)
--- PASS: TestRouter/Reqs_matched_correctly/POST_/api (0.00s)
--- PASS: TestRouter/Reqs_matched_correctly/POST_/api/ (0.00s)
--- PASS: TestRouter/Reqs_matched_correctly/DELETE_/api (0.00s)
--- PASS: TestRouter/Reqs_matched_correctly/GET_/api/fake-id (0.00s)
--- PASS: TestRouter/Reqs_matched_correctly/GET_/api/fake-id/bla (0.00s)
--- PASS: TestRouter/Reqs_matched_correctly/GET_/api/fake-id/stuff/faked-fake (0.00s)
--- PASS: TestRouter/Reqs_execute_correctly (0.00s)
--- PASS: TestRouter/Reqs_execute_correctly/POST_/api_without_auth (0.00s)
--- PASS: TestRouter/Reqs_execute_correctly/POST_/api_with_auth (0.00s)
--- PASS: TestRouter/Reqs_execute_correctly/GET_/api (0.00s)
--- PASS: TestRouter/Overlapping_routes (0.00s)
PASS
ok  	github.com/seantcanavan/lambda_jwt_router/lambda_router	0.005s
```
