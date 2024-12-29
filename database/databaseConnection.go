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

var Client *mongo.Client

// DBinstance initializes and returns a MongoDB client instance
func DBinstance() (*mongo.Client, error) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file, proceeding with environment variables")
	}

	// Retrieve the MONGO_URL environment variable
	MongoDb := os.Getenv("MONGODB_URL")
	if MongoDb == "" {
		return nil, fmt.Errorf("MONGO_URL environment variable is not set")
	}

	// Create a new client and establish a connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoDb))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database to ensure the connection is active
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	log.Println("Connected to MongoDB!")
	return client, nil
}

// OpenCollection creates a reference to a specific collection
func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	return client.Database("gambl").Collection(collectionName)
}

// Initialize the global Client variable
func init() {
	var err error
	Client, err = DBinstance()
	if err != nil {
		log.Fatalf("Could not initialize MongoDB client: %v", err)
	}
}
