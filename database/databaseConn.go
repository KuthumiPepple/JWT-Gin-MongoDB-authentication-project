package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBinstance() *mongo.Client {
	mongoDbUri := os.Getenv("MONGODB_URL")
	if mongoDbUri == "" {
		log.Fatal("MONGODB_URL is empty. Provide value for the variable in the .env file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoDbUri))
	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return client
}

func OpenCollection(dbName, collectionName string) *mongo.Collection {
	client := DBinstance()
	collection := client.Database(dbName).Collection(collectionName)
	return collection
}
