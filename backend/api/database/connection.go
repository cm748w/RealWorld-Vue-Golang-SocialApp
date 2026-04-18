package database

import (
	"context"
	"fmt"
	"os"
	"time"

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
	return nil
}
