// Package lcom contains common vars, consts, and types that are used throughout the program. Additionally, users who download
// and install this library should investigate the contents of this package to get the most out of the library and its functionality.
package lcom

import (
	"context"
	"errors"
	"github.com/aws/aws-lambda-go/events"
)

// Use these const values to populate your own custom claim values

const JWTClaimAudienceKey = "aud"
const JWTClaimEmailKey = "email"
const JWTClaimExpiresAtKey = "exp"
const JWTClaimFirstNameKey = "firstName"
const JWTClaimFullNameKey = "fullName"
const JWTClaimIDKey = "jti"
const JWTClaimIssuedAtKey = "iat"
const JWTClaimIssuerKey = "iss"
const JWTClaimLevelKey = "level"
const JWTClaimNotBeforeKey = "nbf"
const JWTClaimSubjectKey = "sub"
const JWTClaimUserTypeKey = "userType"

// Use these values to get / set the appropriate environment variables for CORS

const CORSHeadersEnvKey = "LAMBDA_JWT_ROUTER_CORS_HEADERS"
const CORSHeadersHeaderKey = "Access-Control-Allow-Headers"
const CORSMethodsEnvKey = "LAMBDA_JWT_ROUTER_CORS_METHODS"
const CORSMethodsHeaderKey = "Access-Control-Allow-Methods"
const CORSOriginEnvKey = "LAMBDA_JWT_ROUTER_CORS_ORIGIN"
const CORSOriginHeaderKey = "Access-Control-Allow-Origin"

// Use these values for general environment configuration

const HMACSecretEnvKey = "LAMBDA_JWT_ROUTER_HMAC_SECRET"
const NoCORS = "LAMBDA_JWT_ROUTER_NO_CORS"

// ContentTypeKey exists because "Content-Type" is not in the http std lib for some reason...
const ContentTypeKey = "Content-Type"

// Use these values to get/set values in the global context

const LambdaContextMethodKey = "method"
const LambdaContextMultiParamsKey = "multiParams"
const LambdaContextPathKey = "path"
const LambdaContextPathParamsKey = "pathParams"
const LambdaContextQueryParamsKey = "queryParams"
const LambdaContextRequestIDKey = "requestId"

var ErrMarshalMapClaims = errors.New("unable to Marshal map claims: %w")
var ErrNoAuthorizationHeader = errors.New("no Authorization header value set: %w")
var ErrNoBearerPrefix = errors.New("missing 'Bearer ' prefix for Authorization header value: %w")
var ErrVerifyJWT = errors.New("unable to verify JWT to retrieve claims. try logging in again to ensure it is not expired: %w")
var ErrBadClaimsObject = errors.New("lambda_jwt_router: the provided object to extract claims into is not compatible with the default claim set and its types: %w")
var ErrUnableToSignToken = errors.New("lambda_jwt_router: the provided claims were unable to be signed: %w")
var ErrInvalidJWT = errors.New("lambda_jwt_router: the provided JWT is invalid: %w")
var ErrInvalidToken = errors.New("lambda_jwt_router: the provided jwt was unable to be parsed into a token: %w")
var ErrInvalidTokenClaims = errors.New("lambda_jwt_router: the provided jwt was unable to be parsed for map claims: %w")
var ErrUnsupportedSigningMethod = errors.New("lambda_jwt_router:the provided signing method is unsupported. HMAC only allowed: %w")

// Handler is a lambda request handler function. It takes in the context value created by API Gateway when proxying to
// AWS Lambda in addition to the events.APIGatewayProxyRequest event itself. This request object is created by API Gateway
// and has all the requisite fields set automagically. When running this library locally, it will automagically fill in
// the lambda request object manually so that your local code will mirror exactly what locally when it's run remotely
// in the AWS Lambda environment. An error value should  be returned as well. Note that your application's errors should never
// set the error value here but instead return an error message or type embedded in the body of the response. This error is for
// critical stack crashes / go crashes.
// Example:
//
//	 func listSomethings(ctx context.Context, req events.APIGatewayProxyRequest) (res events.APIGatewayProxyResponse err error) {
//	     // parse input
//	     var input listSomethingsInput
//	     err = UnmarshalReq(req, false, &input)
//	     if err != nil {
//	         return ErrorRes(err)
//	     }
//
//	     // call some business logic that generates an output struct
//	     // ...
//
//	     return SuccessRes(result)
//	}
type Handler func(context.Context, events.APIGatewayProxyRequest) (
	events.APIGatewayProxyResponse,
	error,
)

// Middleware is a function that receives a handler function (the next function
// in the chain, possibly another middleware or the actual handler matched for
// a req), and returns a handler function. These functions are quite similar
// to HTTP middlewares in other libraries.
//
// Example middleware that logs all reqs:
//
//	func loggerMiddleware(next lmdrouter.Handler) lmdrouter.Handler {
//	    return func(ctx context.Context, req events.APIGatewayProxyRequest) (
//	        res events.APIGatewayProxyResponse,
//	        err error,
//	    ) {
//	        format := "[%s] [%s %s] [%d]%s"
//	        level := "INF"
//	        var code int
//	        var extra string
//
//	        res, err = next(ctx, req)
//	        if err != nil {
//	            level = "ERR"
//	            code = http.StatusInternalServerError
//	            extra = " " + err.ErrorRes()
//	        } else {
//	            code = res.StatusCode
//	            if code >= 400 {
//	                level = "ERR"
//	            }
//	        }
//
//	        log.Printf(format, level, req.HTTPMethod, req.Path, code, extra)
//
//	        return res, err
//	    }
//	}
type Middleware func(Handler) Handler
