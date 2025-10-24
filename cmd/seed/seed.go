package main

import (
	"context"
	"log"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/models"
	"github.com/flowkit/backend/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	// Load environment variables
	config.LoadEnv()

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := config.ConnectDB(ctx)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer client.Disconnect(context.Background())

	config.InitDB(client)
	log.Println("✅ Connected to MongoDB")

	// Clear existing data
	log.Println("🧹 Clearing existing data...")
	config.UsersCollection.DeleteMany(ctx, bson.M{})
	config.LeavesCollection.DeleteMany(ctx, bson.M{})

	// Create test users
	log.Println("👥 Creating test users...")

	users := []models.User{
		{
			FirstName:  "John",
			LastName:   "Doe",
			Email:      "john.employee@flowkit.com",
			Department: "Engineering",
			Role:       "employee",
			LeaveBalance: models.LeaveBalance{
				Total:     28,
				Available: 26,
				Used:      2,
			},
		},
		{
			FirstName:  "Jane",
			LastName:   "Smith",
			Email:      "jane.hod@flowkit.com",
			Department: "Engineering",
			Role:       "HOD",
			LeaveBalance: models.LeaveBalance{
				Total:     28,
				Available: 28,
				Used:      0,
			},
		},
		{
			FirstName:  "Michael",
			LastName:   "Johnson",
			Email:      "michael.hr@flowkit.com",
			Department: "Human Resources",
			Role:       "HR",
			LeaveBalance: models.LeaveBalance{
				Total:     28,
				Available: 28,
				Used:      0,
			},
		},
		{
			FirstName:  "Sarah",
			LastName:   "Williams",
			Email:      "sarah.ged@flowkit.com",
			Department: "Management",
			Role:       "GED",
			LeaveBalance: models.LeaveBalance{
				Total:     28,
				Available: 28,
				Used:      0,
			},
		},
		{
			FirstName:  "Admin",
			LastName:   "User",
			Email:      "admin@flowkit.com",
			Department: "Management",
			Role:       "admin",
			LeaveBalance: models.LeaveBalance{
				Total:     28,
				Available: 28,
				Used:      0,
			},
		},
		{
			FirstName:  "Alice",
			LastName:   "Brown",
			Email:      "alice.employee@flowkit.com",
			Department: "Marketing",
			Role:       "employee",
			LeaveBalance: models.LeaveBalance{
				Total:     20,
				Available: 20,
				Used:      0,
			},
		},
		{
			FirstName:  "Bob",
			LastName:   "Davis",
			Email:      "bob.employee@flowkit.com",
			Department: "Finance",
			Role:       "employee",
			LeaveBalance: models.LeaveBalance{
				Total:     20,
				Available: 19,
				Used:      1,
			},
		},
	}

	// Hash password and insert users
	password := "password123"
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	for i := range users {
		users[i].Password = hashedPassword
		staffID, _ := utils.GenerateStaffID(ctx)
		users[i].StaffID = staffID
		users[i].CreatedAt = time.Now()
		users[i].UpdatedAt = time.Now()

		result, err := config.UsersCollection.InsertOne(ctx, users[i])
		if err != nil {
			log.Printf("Failed to insert user %s: %v", users[i].Email, err)
		} else {
			users[i].ID = result.InsertedID.(primitive.ObjectID)
			log.Printf("✅ Created user: %s (%s) - %s", users[i].Email, users[i].Role, users[i].StaffID)
		}
	}

	// Create sample leave requests
	log.Println("\n📝 Creating sample leave requests...")

	// Leave 1 - Pending at HOD stage
	leave1 := models.Leave{
		Employee:     users[0].ID, // John Doe (employee)
		LeaveType:    "Annual Leave",
		FromDate:     time.Now().AddDate(0, 0, 10),
		ToDate:       time.Now().AddDate(0, 0, 14),
		TotalDays:    5,
		Reliever:     users[5].ID, // Alice Brown
		Reason:       "Family vacation",
		Status:       "Pending",
		Stage:        1,
		IsEditable:   true,
		ApprovalFlow: []models.ApprovalStep{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Leave 2 - Approved by HOD, pending at HR
	leave2 := models.Leave{
		Employee:   users[0].ID, // John Doe
		LeaveType:  "Sick Leave",
		FromDate:   time.Now().AddDate(0, 0, 20),
		ToDate:     time.Now().AddDate(0, 0, 22),
		TotalDays:  3,
		Reliever:   users[5].ID,
		Reason:     "Medical appointment",
		Status:     "Pending",
		Stage:      2,
		IsEditable: false,
		ApprovalFlow: []models.ApprovalStep{
			{
				Approver: users[1].ID, // Jane Smith (HOD)
				Role:     "HOD",
				Status:   "Approved",
				Comments: "Approved by HOD",
				Date:     time.Now().AddDate(0, 0, -2),
			},
		},
		CreatedAt: time.Now().AddDate(0, 0, -2),
		UpdatedAt: time.Now().AddDate(0, 0, -2),
	}

	// Leave 3 - Fully approved
	leave3 := models.Leave{
		Employee:   users[6].ID, // Bob Davis
		LeaveType:  "Annual Leave",
		FromDate:   time.Now().AddDate(0, 0, 5),
		ToDate:     time.Now().AddDate(0, 0, 7),
		TotalDays:  3,
		Reliever:   users[5].ID,
		Reason:     "Personal matters",
		Status:     "Approved",
		Stage:      3,
		IsEditable: false,
		ApprovalFlow: []models.ApprovalStep{
			{
				Approver: users[1].ID, // HOD
				Role:     "HOD",
				Status:   "Approved",
				Comments: "Approved",
				Date:     time.Now().AddDate(0, 0, -5),
			},
			{
				Approver: users[2].ID, // HR
				Role:     "HR",
				Status:   "Approved",
				Comments: "Approved",
				Date:     time.Now().AddDate(0, 0, -3),
			},
			{
				Approver: users[3].ID, // GED
				Role:     "GED",
				Status:   "Approved",
				Comments: "Approved",
				Date:     time.Now().AddDate(0, 0, -1),
			},
		},
		CreatedAt: time.Now().AddDate(0, 0, -5),
		UpdatedAt: time.Now().AddDate(0, 0, -1),
	}

	// Leave 4 - Rejected
	leave4 := models.Leave{
		Employee:   users[5].ID, // Alice Brown
		LeaveType:  "Emergency Leave",
		FromDate:   time.Now().AddDate(0, 0, 15),
		ToDate:     time.Now().AddDate(0, 0, 17),
		TotalDays:  3,
		Reliever:   users[0].ID,
		Reason:     "Emergency travel",
		Status:     "Rejected",
		Stage:      1,
		IsEditable: false,
		ApprovalFlow: []models.ApprovalStep{
			{
				Approver: users[1].ID, // HOD
				Role:     "HOD",
				Status:   "Rejected",
				Comments: "Insufficient notice period",
				Date:     time.Now().AddDate(0, 0, -1),
			},
		},
		CreatedAt: time.Now().AddDate(0, 0, -2),
		UpdatedAt: time.Now().AddDate(0, 0, -1),
	}

	// Insert leave requests
	leaves := []models.Leave{leave1, leave2, leave3, leave4}
	for _, leave := range leaves {
		_, err := config.LeavesCollection.InsertOne(ctx, leave)
		if err != nil {
			log.Printf("Failed to insert leave: %v", err)
		} else {
			log.Printf("✅ Created leave request: %s - %d days (%s)", leave.LeaveType, leave.TotalDays, leave.Status)
		}
	}

	log.Println("\n✅ Database seeded successfully!")
	log.Println("\n📋 Test Users:")
	log.Println("   Employee: john.employee@flowkit.com / password123")
	log.Println("   HOD:      jane.hod@flowkit.com / password123")
	log.Println("   HR:       michael.hr@flowkit.com / password123")
	log.Println("   GED:      sarah.ged@flowkit.com / password123")
	log.Println("   Admin:    admin@flowkit.com / password123")
}
