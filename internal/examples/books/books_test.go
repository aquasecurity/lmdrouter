package books

import (
	"context"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/seantcanavan/lambda_jwt_router/internal/examples/database"
	"github.com/seantcanavan/lambda_jwt_router/internal/util"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var textCtx = context.Background()

func TestMain(m *testing.M) {
	setup()
	m.Run()
	tearDown()
}

func setup() {
	err := godotenv.Load("../../../.env")
	if err != nil {
		panic(fmt.Sprintf("Unable to load .env file: %s", err))
	}

	time.Local = time.UTC
	database.Connect()
}

func tearDown() {
	database.Disconnect()
}

func TestBooks(t *testing.T) {
	t.Run("verify Create is working as expected", func(t *testing.T) {
		cReq := &CreateReq{
			Author: util.GenerateRandomString(10),
			Pages:  util.GenerateRandomInt(500, 5000),
			Title:  util.GenerateRandomString(10),
		}

		createdBook, err := Create(textCtx, cReq)
		require.NoError(t, err)

		require.Equal(t, cReq.Author, createdBook.Author)
		require.Equal(t, cReq.Pages, createdBook.Pages)
		require.Equal(t, cReq.Title, createdBook.Title)
		require.False(t, createdBook.ID.IsZero())

		t.Run("verify Get is working as expected", func(t *testing.T) {
			gReq := &GetReq{ID: createdBook.ID}

			book, err := Get(textCtx, gReq)
			require.NoError(t, err)

			require.Equal(t, createdBook.Title, book.Title)
			require.Equal(t, createdBook.Author, book.Author)
			require.Equal(t, createdBook.ID, book.ID)
			require.Equal(t, createdBook.Pages, book.Pages)
		})

		t.Run("verify Update is working as expected", func(t *testing.T) {
			uReq := &UpdateReq{
				Author: util.GenerateRandomString(10),
				ID:     createdBook.ID,
				Pages:  util.GenerateRandomInt(500, 5000),
				Title:  util.GenerateRandomString(10),
			}

			updatedBook, err := Update(textCtx, uReq)
			require.NoError(t, err)

			require.Equal(t, uReq.Author, updatedBook.Author)
			require.Equal(t, uReq.Title, updatedBook.Title)
			require.Equal(t, uReq.Pages, updatedBook.Pages)
			require.Equal(t, uReq.ID, createdBook.ID)

			t.Run("verify Delete is working as expected", func(t *testing.T) {
				dReq := &DeleteReq{ID: createdBook.ID}

				deletedBook, err := Delete(textCtx, dReq)
				require.NoError(t, err)

				require.Equal(t, updatedBook.Title, deletedBook.Title)
				require.Equal(t, updatedBook.Author, deletedBook.Author)
				require.Equal(t, updatedBook.ID, deletedBook.ID)
				require.Equal(t, updatedBook.Pages, deletedBook.Pages)
			})
		})
	})
}
