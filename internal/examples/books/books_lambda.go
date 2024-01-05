package books

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/seantcanavan/lambda_jwt_router/lreq"
	"github.com/seantcanavan/lambda_jwt_router/lres"
	"net/http"
)

func CreateLambda(ctx context.Context, lambdaReq events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cReq := &CreateReq{}
	err := lreq.UnmarshalReq(lambdaReq, true, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	book, err := Create(ctx, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	return lres.SuccessRes(book)
}

func DeleteLambda(ctx context.Context, lambdaReq events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cReq := &DeleteReq{}
	err := lreq.UnmarshalReq(lambdaReq, false, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	book, err := Delete(ctx, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	return lres.SuccessRes(book)
}

func GetLambda(ctx context.Context, lambdaReq events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cReq := &GetReq{}
	err := lreq.UnmarshalReq(lambdaReq, false, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	book, err := Get(ctx, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	return lres.SuccessRes(book)
}

func UpdateLambda(ctx context.Context, lambdaReq events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	cReq := &UpdateReq{}
	err := lreq.UnmarshalReq(lambdaReq, true, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	book, err := Update(ctx, cReq)
	if err != nil {
		return lres.StatusAndErrorRes(http.StatusInternalServerError, err)
	}

	return lres.SuccessRes(book)
}
