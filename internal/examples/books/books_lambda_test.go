package books

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/seantcanavan/lambda_jwt_router/lreq"
	"github.com/seantcanavan/lambda_jwt_router/lres"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestBooksLambda(t *testing.T) {
	t.Run("verify CreateLambda is working as expected", func(t *testing.T) {
		cReq := &CreateReq{
			Author: util.GenerateRandomString(10),
			Pages:  util.GenerateRandomInt(500, 5000),
			Title:  util.GenerateRandomString(10),
		}

		createRes, err := CreateLambda(textCtx, lreq.MarshalReq(cReq))
		require.NoError(t, err)

		createdBook := &Book{}
		err = lres.UnmarshalRes(createRes, createdBook)
		require.NoError(t, err)

		require.Equal(t, cReq.Author, createdBook.Author)
		require.Equal(t, cReq.Pages, createdBook.Pages)
		require.Equal(t, cReq.Title, createdBook.Title)
		require.False(t, createdBook.ID.IsZero())

		t.Run("verify GetLambda is working as expected", func(t *testing.T) {
			getRes, err := GetLambda(textCtx, events.APIGatewayProxyRequest{PathParameters: map[string]string{"id": createdBook.ID.Hex()}})
			require.NoError(t, err)

			gotBook := &Book{}

			err = lres.UnmarshalRes(getRes, gotBook)
			require.NoError(t, err)

			require.Equal(t, createdBook.Title, gotBook.Title)
			require.Equal(t, createdBook.Author, gotBook.Author)
			require.Equal(t, createdBook.ID, gotBook.ID)
			require.Equal(t, createdBook.Pages, gotBook.Pages)
		})

		t.Run("verify UpdateLambda is working as expected", func(t *testing.T) {
			uReq := &UpdateReq{
				Author: util.GenerateRandomString(10),
				Pages:  util.GenerateRandomInt(500, 5000),
				Title:  util.GenerateRandomString(10),
			}

			lambdaReq := lreq.MarshalReq(uReq)
			lambdaReq.PathParameters = map[string]string{"id": createdBook.ID.Hex()}

			updateRes, err := UpdateLambda(textCtx, lambdaReq)
			require.NoError(t, err)

			updatedBook := &Book{}
			err = lres.UnmarshalRes(updateRes, updatedBook)
			require.NoError(t, err)

			require.Equal(t, uReq.Author, updatedBook.Author)
			require.Equal(t, uReq.Title, updatedBook.Title)
			require.Equal(t, uReq.Pages, updatedBook.Pages)
			require.Equal(t, createdBook.ID, updatedBook.ID)

			t.Run("verify DeleteLambda is working as expected", func(t *testing.T) {
				deleteRes, err := DeleteLambda(textCtx, events.APIGatewayProxyRequest{PathParameters: map[string]string{"id": createdBook.ID.Hex()}})
				require.NoError(t, err)

				deletedBook := &Book{}
				err = lres.UnmarshalRes(deleteRes, deletedBook)
				require.NoError(t, err)

				require.Equal(t, updatedBook.Title, deletedBook.Title)
				require.Equal(t, updatedBook.Author, deletedBook.Author)
				require.Equal(t, updatedBook.ID, deletedBook.ID)
				require.Equal(t, updatedBook.Pages, deletedBook.Pages)
			})
		})
	})

}
