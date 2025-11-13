package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

	// Check if admin already exists
	collection := config.DB.Collection("users")
	ctx2, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var existingUser models.User
	findErr := collection.FindOne(ctx2, bson.M{"email": "admin@flowkit.com"}).Decode(&existingUser)
	if findErr == nil {
		fmt.Println("Admin user already exists!")
		fmt.Printf("Email: admin@flowkit.com\n")
		fmt.Printf("Name: %s %s\n", existingUser.FirstName, existingUser.LastName)
		fmt.Printf("Role: %s\n", existingUser.Role)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Error hashing password:", err)
	}

	// Create admin user
	adminUser := models.User{
		ID:         primitive.NewObjectID(),
		FirstName:  "Admin",
		LastName:   "User",
		Email:      "admin@flowkit.com",
		Password:   string(hashedPassword),
		StaffID:    "ADMIN001",
		Department: "Administration",
		Role:       "admin",
		IsHOD:      false,
		LeaveBalance: models.LeaveBalance{
			Total:     28,
			Available: 28,
			Used:      0,
		},
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Insert admin user
	_, err = collection.InsertOne(ctx2, adminUser)
	if err != nil {
		log.Fatal("Error creating admin user:", err)
	}

	fmt.Println("✅ Admin user created successfully!")
	fmt.Println("\n📋 Login Credentials:")
	fmt.Println("Email: admin@flowkit.com")
	fmt.Println("Password: Admin@123")
	fmt.Println("\nYou can now use these credentials to log in and test the admin functionality.")
}
