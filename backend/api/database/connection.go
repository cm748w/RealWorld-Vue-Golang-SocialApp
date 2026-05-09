package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var DB *mongo.Database

func Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoUri := os.Getenv("MONGO_URI")
	if mongoUri == "" {
		mongoUri = "mongodb://localhost:27017"
	}

	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoUri))

	if err != nil {
		fmt.Printf("error connecting to db: %v\n", err)
		return err
	}

	fmt.Println("Connected to MongoDB")
	DB = Client.Database("social")

	// Enforce one unread counter document per (receiver, sender) pair.
	unreadCollection := DB.Collection("unreadmessages")
	_, err = unreadCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "mainUserId", Value: 1}, {Key: "otherUserId", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_main_other_unread"),
	})
	if err != nil {
		fmt.Printf("warning: failed to create unreadmessages unique index: %v\n", err)
	}

	return nil
}
