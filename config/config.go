package config

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	DB               *mongo.Database
	UsersCollection  *mongo.Collection
	LeavesCollection *mongo.Collection
)

// LoadEnv loads environment variables from .env file
func LoadEnv() {
	file, err := os.Open(".env")
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			os.Setenv(key, value)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading .env file: %v", err)
	}
}

// ConnectDB connects to MongoDB
func ConnectDB(ctx context.Context) (*mongo.Client, error) {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		log.Fatal("MONGODB_URI environment variable is not set")
	}

	// Create client options
	clientOptions := options.Client().ApplyURI(mongoURI)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Printf("📦 Database: %s", client.Database("flowkit_leave_management").Name())
	return client, nil
}

// InitDB initializes database collections
func InitDB(client *mongo.Client) {
	DB = client.Database("flowkit_leave_management")
	UsersCollection = DB.Collection("users")
	LeavesCollection = DB.Collection("leaves")
}
