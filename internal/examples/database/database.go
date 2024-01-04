package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"time"
)

var Client *mongo.Client
var Database *mongo.Database

var BooksColl *mongo.Collection

func Connect() {
	name := "lambda_jwt_router"
	connectionString := os.Getenv("CONNECTION_STRING")
	ctx := context.Background()

	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		panic(fmt.Sprintf("mongo.NewClient failed with error %s", err))
	}

	Database = Client.Database(name)

	BooksColl = Database.Collection("books")

	createIndexesOptions := options.CreateIndexesOptions{
		CommitQuorum: "majority",
	}

	_, err = BooksColl.Indexes().CreateMany(ctx, getBookIndexes(), &createIndexesOptions)
	if err != nil {
		panic(fmt.Sprintf("BooksColl.Indexes().CreateMany failed with error %s", err))
	}
}

func getBookIndexes() []mongo.IndexModel {
	var indexes []mongo.IndexModel

	// create an index on fullName, so we can search by name
	authorTitleUniqueIndex := mongo.IndexModel{
		Keys: bson.D{{"author", 1}, {"title", 1}},
		Options: &options.IndexOptions{
			Unique: func() *bool { a := true; return &a }(),
		},
	}
	indexes = append(indexes, authorTitleUniqueIndex)

	return indexes
}

func Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err := Client.Disconnect(ctx)
	if err != nil {
		panic(fmt.Sprintf("Client.Disconnect(ctx) failed with error %s", err))
	}

	defer cancel()
}
