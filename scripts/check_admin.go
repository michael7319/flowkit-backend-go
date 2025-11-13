package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
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

	// Find admin user
	collection := config.DB.Collection("users")
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var adminUser models.User
	err = collection.FindOne(ctx2, bson.M{"email": "admin@flowkit.com"}).Decode(&adminUser)
	if err != nil {
		log.Fatal("Admin user not found:", err)
	}

	fmt.Println("📋 Current Admin User in Database:")
	fmt.Printf("ID: %s\n", adminUser.ID.Hex())
	fmt.Printf("Name: %s %s\n", adminUser.FirstName, adminUser.LastName)
	fmt.Printf("Email: %s\n", adminUser.Email)
	fmt.Printf("Role: %s\n", adminUser.Role)
	fmt.Printf("Department: %s\n", adminUser.Department)
	fmt.Printf("Active: %v\n", adminUser.IsActive)
	fmt.Printf("Staff ID: %s\n", adminUser.StaffID)

	// Test password
	testPassword := "Admin@123"
	err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(testPassword))
	if err == nil {
		fmt.Printf("\n✅ Password 'Admin@123' is CORRECT\n")
	} else {
		fmt.Printf("\n❌ Password 'Admin@123' does NOT match stored hash\n")
		fmt.Println("Updating password to 'Admin@123'...")

		// Hash the new password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal("Error hashing password:", err)
		}

		// Update password
		update := bson.M{
			"$set": bson.M{
				"password":  string(hashedPassword),
				"updatedAt": time.Now(),
			},
		}

		_, err = collection.UpdateOne(ctx2, bson.M{"email": "admin@flowkit.com"}, update)
		if err != nil {
			log.Fatal("Error updating password:", err)
		}

		fmt.Println("✅ Password updated successfully!")
	}

	fmt.Println("\n🔐 Login Credentials:")
	fmt.Println("Email: admin@flowkit.com")
	fmt.Println("Password: Admin@123")
}
