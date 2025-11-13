package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/models"
	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Initialize database connection
	ctx := context.Background()
	client, err := config.ConnectDB(ctx)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	// Initialize collections
	config.InitDB(client)

	// Find and activate admin user
	collection := config.DB.Collection("users")
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update admin user to set isActive to true
	filter := bson.M{"email": "admin@flowkit.com"}
	update := bson.M{
		"$set": bson.M{
			"isActive":  true,
			"updatedAt": time.Now(),
		},
	}

	result, err := collection.UpdateOne(ctx2, filter, update)
	if err != nil {
		log.Fatal("Error activating admin user:", err)
	}

	if result.MatchedCount == 0 {
		fmt.Println("❌ Admin user not found!")
		return
	}

	// Fetch and display the updated user
	var adminUser models.User
	err = collection.FindOne(ctx2, filter).Decode(&adminUser)
	if err != nil {
		log.Fatal("Error fetching admin user:", err)
	}

	fmt.Println("✅ Admin user activated successfully!")
	fmt.Println("\n📋 Admin Account Details:")
	fmt.Printf("Name: %s %s\n", adminUser.FirstName, adminUser.LastName)
	fmt.Printf("Email: %s\n", adminUser.Email)
	fmt.Printf("Role: %s\n", adminUser.Role)
	fmt.Printf("Active: %v\n", adminUser.IsActive)
	fmt.Println("\n🔐 Login Credentials:")
	fmt.Println("Email: admin@flowkit.com")
	fmt.Println("Password: Admin@123")
	fmt.Println("\nYou can now log in with these credentials!")
}
