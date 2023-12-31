package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error while loading '.env' file")
	}
	mongoDbUri := os.Getenv("MONGODB_URL")
	if mongoDbUri == "" {
		log.Fatal("MONGODB_URL is empty. Provide value for the variable in the .env file")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoDbUri))
	if err != nil {
		log.Panic(err)
	}
	if err = client.Ping(ctx, nil); err != nil {
		log.Panic(err)
	}
	fmt.Println("Connected to MongoDB!")
}

func OpenCollection(collectionName string) *mongo.Collection {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "usersDB"
	}
	collection := client.Database(dbName).Collection(collectionName)
	return collection
}
