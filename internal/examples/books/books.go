package books

import (
	"context"
	"errors"
	"github.com/seantcanavan/lambda_jwt_router/internal/examples/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Book struct {
	Author string             `bson:"author,omitempty" json:"author,omitempty"`
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Pages  int                `bson:"pages,omitempty" json:"pages,omitempty"`
	Title  string             `bson:"title,omitempty" json:"title,omitempty"`
}

// Valid is a naive implementation of a validator for the Book CreateReq
// In production you should use something like https://github.com/go-playground/validator
// to perform validation in-line with unmarshalling for a more cohesive experience
func (cReq *CreateReq) Valid() error {
	if cReq.Author == "" {
		return errors.New("author is required")
	}

	if cReq.Title == "" {
		return errors.New("title is required")
	}

	if cReq.Pages == 0 {
		return errors.New("pages is required and must be non-zero")
	}

	return nil
}

type CreateReq struct {
	Author string `json:"author,omitempty"`
	Title  string `json:"title,omitempty"`
	Pages  int    `json:"pages"`
}

func Create(ctx context.Context, cReq *CreateReq) (*Book, error) {
	if err := cReq.Valid(); err != nil {
		return nil, err
	}

	book := &Book{
		Author: cReq.Author,
		ID:     primitive.NewObjectID(),
		Pages:  cReq.Pages,
		Title:  cReq.Title,
	}

	_, err := database.BooksColl.InsertOne(ctx, book)
	if err != nil {
		return nil, err
	}

	return book, nil

}

// Valid is a naive implementation of a validator for the Book DeleteReq
// In production you should use something like https://github.com/go-playground/validator
// to perform validation in-line with unmarshalling for a more cohesive experience
func (dReq *DeleteReq) Valid() error {
	if dReq.ID.IsZero() {
		return errors.New("id cannot be zero")
	}

	return nil
}

type DeleteReq struct {
	ID primitive.ObjectID `lambda:"path.id" json:"id,omitempty"`
}

func Delete(ctx context.Context, dReq *DeleteReq) (*Book, error) {
	if err := dReq.Valid(); err != nil {
		return nil, err
	}

	singleRes := database.BooksColl.FindOneAndDelete(ctx, bson.M{"_id": dReq.ID})
	if singleRes.Err() != nil {
		return nil, singleRes.Err()
	}

	book := &Book{}
	err := singleRes.Decode(book)
	if err != nil {
		return nil, err
	}

	return book, nil
}

// Valid is a naive implementation of a validator for the Book GetReq
// In production you should use something like https://github.com/go-playground/validator
// to perform validation in-line with unmarshalling for a more cohesive experience
func (gReq *GetReq) Valid() error {
	if gReq.ID.IsZero() {
		return errors.New("id cannot be zero")
	}

	return nil
}

type GetReq struct {
	ID primitive.ObjectID `lambda:"path.id" json:"id,omitempty"`
}

func Get(ctx context.Context, gReq *GetReq) (*Book, error) {
	if err := gReq.Valid(); err != nil {
		return nil, err
	}

	singleRes := database.BooksColl.FindOne(ctx, bson.M{"_id": gReq.ID})
	if singleRes.Err() != nil {
		return nil, singleRes.Err()
	}

	book := &Book{}
	err := singleRes.Decode(book)
	if err != nil {
		return nil, err
	}

	return book, nil
}

// Valid is a naive implementation of a validator for the Book UpdateReq
// In production you should use something like https://github.com/go-playground/validator
// to perform validation in-line with unmarshalling for a more cohesive experience
func (uReq *UpdateReq) Valid() error {
	if uReq.Author == "" {
		return errors.New("author is required")
	}

	if uReq.ID.IsZero() {
		return errors.New("id cannot be zero")
	}

	if uReq.Pages == 0 {
		return errors.New("pages is required and must be non-zero")
	}

	if uReq.Title == "" {
		return errors.New("title is required")
	}

	return nil
}

type UpdateReq struct {
	Author string             `json:"author,omitempty"`
	ID     primitive.ObjectID `lambda:"path.id" json:"id,omitempty"`
	Pages  int                `json:"pages"`
	Title  string             `json:"title,omitempty"`
}

func Update(ctx context.Context, uReq *UpdateReq) (*Book, error) {
	if err := uReq.Valid(); err != nil {
		return nil, err
	}

	singleRes := database.BooksColl.FindOneAndUpdate(ctx, bson.M{"_id": uReq.ID},
		bson.M{
			"$set": bson.M{
				"author": uReq.Author,
				"title":  uReq.Title,
				"pages":  uReq.Pages,
			},
		},
		&options.FindOneAndUpdateOptions{ReturnDocument: func() *options.ReturnDocument { a := options.After; return &a }()},
	)
	if singleRes.Err() != nil {
		return nil, singleRes.Err()
	}

	book := &Book{}
	err := singleRes.Decode(book)
	if err != nil {
		return nil, err
	}

	return book, nil
}
