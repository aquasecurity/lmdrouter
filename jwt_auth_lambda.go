package lmdrouter

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
)

// DecodeJWTMiddleware attempts to parse a Json Web Token from the req's "Authorization"
// header. If the Authorization header is missing, or does not contain a valid Json Web Token
// (JWT) then an error message and appropriate HTTP status code will be returned
func DecodeJWTMiddleware(next Handler) Handler {
	return func(ctx context.Context, req events.APIGatewayProxyRequest) (
		res events.APIGatewayProxyResponse,
		err error,
	) {
		if req.HTTPMethod == "OPTIONS" {
			// immediately return success for options calls for CORS reqs
			return MarshalResponse(200, nil, nil)
		}

		adminClaim, httpStatus, err := JwtAuth(req.Headers)
		if err != nil {
			return HandleHTTPError(httpStatus, err)
		}

		ctx = context.WithValue(ctx, "adminFullName", adminClaim.FullName)
		ctx = context.WithValue(ctx, "adminID", adminClaim.ID)
		ctx = context.WithValue(ctx, "adminLevel", adminClaim.Level)
		ctx = context.WithValue(ctx, "expirationDate", adminClaim.ExpirationDate)

		res, err = next(ctx, req)

		return res, err
	}
}
