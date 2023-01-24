package jwt

//
//import (
//	"context"
//	"fmt"
//	"github.com/aws/aws-lambda-go/events"
//	"github.com/golang-jwt/jwt"
//	"net/http"
//	"strings"
//)
//
//// DecodeStandardJWTMiddleware attempts to parse a Json Web Token from the req's "Authorization"
//// header. If the Authorization header is missing, or does not contain a valid Json Web Token
//// (JWT) then an error message and appropriate HTTP status code will be returned. If the JWT
//// is correctly set and contains a StandardClaim then the values from tha standard claim
//// will be added to the context object for others to use during their processing.
//func DecodeStandardJWTMiddleware(next Handler) Handler {
//	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
//		res events.APIGatewayProxyResponse,
//		err error,
//	) {
//		if req.HTTPMethod == "OPTIONS" {
//			// immediately return success for options calls for CORS reqs
//			return MarshalResponse(200, nil, nil)
//		}
//
//		mapClaims, httpStatus, err := ExtractJWT(req.Headers)
//		if err != nil {
//			return HandleHTTPError(httpStatus, err)
//		}
//
//		standardClaims, err := ExtractStandardClaims(mapClaims)
//		if err != nil {
//			return HandleHTTPError(http.StatusInternalServerError, err)
//		}
//
//		ctx = context.WithValue(ctx, AudienceKey, standardClaims.Audience)
//		ctx = context.WithValue(ctx, ExpiresAtKey, standardClaims.ExpiresAt)
//		ctx = context.WithValue(ctx, IDKey, standardClaims.Id)
//		ctx = context.WithValue(ctx, IssuedAtKey, standardClaims.IssuedAt)
//		ctx = context.WithValue(ctx, IssuerKey, standardClaims.Issuer)
//		ctx = context.WithValue(ctx, NotBeforeKey, standardClaims.NotBefore)
//		ctx = context.WithValue(ctx, SubjectKey, standardClaims.Subject)
//
//		res, err = next(ctx, req)
//
//		return res, err
//	}
//}
//
//func ExtractJWT(headers map[string]string) (jwt.MapClaims, int, error) {
//	authorizationHeader := headers["Authorization"]
//	if authorizationHeader == "" {
//		return nil, http.StatusUnauthorized, fmt.Errorf("missing Authorization header value")
//	}
//
//	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
//		return nil, http.StatusUnauthorized, fmt.Errorf("missing 'Bearer ' prefix for Authorization header")
//	}
//
//	userJwt := strings.TrimPrefix(authorizationHeader, "Bearer ")
//
//	mapClaims, httpStatus, err := VerifyJWT(userJwt)
//	if err != nil {
//		return nil, httpStatus, err
//	}
//
//	return mapClaims, http.StatusOK, nil
//}
