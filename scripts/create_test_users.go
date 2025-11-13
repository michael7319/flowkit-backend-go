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

type TestUser struct {
	FirstName  string
	LastName   string
	Email      string
	Password   string
	Department string
	Role       string
	IsHOD      bool
	StaffID    string
}

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

	collection := config.DB.Collection("users")
	ctx2, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Define test users
	testUsers := []TestUser{
		{
			FirstName:  "John",
			LastName:   "HOD",
			Email:      "hod@flowkit.com",
			Password:   "Hod@123",
			Department: "Engineering",
			Role:       "hod",
			IsHOD:      true,
			StaffID:    "HOD001",
		},
		{
			FirstName:  "Sarah",
			LastName:   "HR",
			Email:      "hr@flowkit.com",
			Password:   "Hr@123",
			Department: "Human Resources",
			Role:       "hr",
			IsHOD:      false,
			StaffID:    "HR001",
		},
		{
			FirstName:  "Michael",
			LastName:   "GED",
			Email:      "ged@flowkit.com",
			Password:   "Ged@123",
			Department: "Executive",
			Role:       "ged",
			IsHOD:      false,
			StaffID:    "GED001",
		},
		{
			FirstName:  "Emma",
			LastName:   "Employee",
			Email:      "employee@flowkit.com",
			Password:   "Employee@123",
			Department: "Engineering",
			Role:       "employee",
			IsHOD:      false,
			StaffID:    "EMP001",
		},
	}

	fmt.Println("🚀 Creating test users for FlowKit Leave Management System...")
	fmt.Println("===========================================================\n")

	successCount := 0

	for _, testUser := range testUsers {
		// Check if user already exists
		var existingUser models.User
		err := collection.FindOne(ctx2, bson.M{"email": testUser.Email}).Decode(&existingUser)
		if err == nil {
			fmt.Printf("⚠️  User already exists: %s (%s)\n", testUser.Email, testUser.Role)

			// Update password for existing user
			hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
			update := bson.M{
				"$set": bson.M{
					"password":  string(hashedPassword),
					"isActive":  true,
					"updatedAt": time.Now(),
				},
			}
			collection.UpdateOne(ctx2, bson.M{"email": testUser.Email}, update)
			fmt.Printf("   ✅ Password updated and account activated\n\n")
			successCount++
			continue
		}

		// Hash the password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(testUser.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("❌ Error hashing password for %s: %v\n", testUser.Email, err)
			continue
		}

		// Create user
		user := models.User{
			ID:         primitive.NewObjectID(),
			FirstName:  testUser.FirstName,
			LastName:   testUser.LastName,
			Email:      testUser.Email,
			Password:   string(hashedPassword),
			StaffID:    testUser.StaffID,
			Department: testUser.Department,
			Role:       testUser.Role,
			IsHOD:      testUser.IsHOD,
			LeaveBalance: models.LeaveBalance{
				Total:     28,
				Available: 28,
				Used:      0,
			},
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Insert user
		_, err = collection.InsertOne(ctx2, user)
		if err != nil {
			log.Printf("❌ Error creating user %s: %v\n", testUser.Email, err)
			continue
		}

		fmt.Printf("✅ Created: %s %s (%s)\n", testUser.FirstName, testUser.LastName, testUser.Role)
		fmt.Printf("   Email: %s\n", testUser.Email)
		fmt.Printf("   Password: %s\n", testUser.Password)
		fmt.Printf("   Department: %s\n", testUser.Department)
		fmt.Printf("   Staff ID: %s\n", testUser.StaffID)
		if testUser.IsHOD {
			fmt.Printf("   IsHOD: true ⭐\n")
		}
		fmt.Println()
		successCount++
	}

	fmt.Println("===========================================================")
	fmt.Printf("✅ Successfully created/updated %d test users!\n\n", successCount)

	fmt.Println("🔐 LOGIN CREDENTIALS:")
	fmt.Println("====================")
	fmt.Println("\n1️⃣  ADMIN (System Administrator)")
	fmt.Println("   Email: admin@flowkit.com")
	fmt.Println("   Password: Admin@123")
	fmt.Println("   Purpose: Manage all users and system settings")

	fmt.Println("\n2️⃣  HOD (Head of Department)")
	fmt.Println("   Email: hod@flowkit.com")
	fmt.Println("   Password: Hod@123")
	fmt.Println("   Purpose: First stage approval - approve/reject department leaves")

	fmt.Println("\n3️⃣  HR (Human Resources)")
	fmt.Println("   Email: hr@flowkit.com")
	fmt.Println("   Password: Hr@123")
	fmt.Println("   Purpose: Second stage approval - review HOD-approved leaves")

	fmt.Println("\n4️⃣  GED (General Executive Director)")
	fmt.Println("   Email: ged@flowkit.com")
	fmt.Println("   Password: Ged@123")
	fmt.Println("   Purpose: Final approval - review HR-approved leaves")

	fmt.Println("\n5️⃣  EMPLOYEE (Regular Employee)")
	fmt.Println("   Email: employee@flowkit.com")
	fmt.Println("   Password: Employee@123")
	fmt.Println("   Purpose: Request and manage leaves")

	fmt.Println("\n\n🎯 TESTING WORKFLOW:")
	fmt.Println("===================")
	fmt.Println("1. Login as Employee → Submit leave request")
	fmt.Println("2. Login as HOD → Approve employee's leave")
	fmt.Println("3. Login as HR → Approve HOD-approved leave")
	fmt.Println("4. Login as GED → Final approval")
	fmt.Println("5. Login as Employee → View fully approved leave")

	fmt.Println("\n✨ All users are now ready for testing!")
}
