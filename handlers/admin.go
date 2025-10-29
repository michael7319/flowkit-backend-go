package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/middleware"
	"github.com/flowkit/backend/models"
	"github.com/flowkit/backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AdminCreateUserRequest represents admin user creation data
type AdminCreateUserRequest struct {
	FirstName  string `json:"firstName" binding:"required"`
	LastName   string `json:"lastName" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Department string `json:"department" binding:"required"`
	Role       string `json:"role" binding:"required"`
	StaffID    string `json:"staffId,omitempty"`
	TotalLeave int    `json:"totalLeave,omitempty"` // Optional, defaults to 28
	IsActive   *bool  `json:"isActive,omitempty"`   // Optional, defaults to true
}

// AdminUpdateUserRequest represents admin user update data
type AdminUpdateUserRequest struct {
	FirstName  string `json:"firstName,omitempty"`
	LastName   string `json:"lastName,omitempty"`
	Email      string `json:"email,omitempty"`
	Department string `json:"department,omitempty"`
	Role       string `json:"role,omitempty"`
	StaffID    string `json:"staffId,omitempty"`
	IsActive   *bool  `json:"isActive,omitempty"`
}

// AdminUpdateLeaveBalanceRequest represents leave balance adjustment
type AdminUpdateLeaveBalanceRequest struct {
	Total     *int `json:"total,omitempty"`
	Available *int `json:"available,omitempty"`
	Used      *int `json:"used,omitempty"`
}

// AdminCreateUser creates a new user account (admin only)
func AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	// Validate department
	if !models.IsValidDepartment(req.Department) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid department. Valid departments: " +
				"VAS, VOICE, ACCOUNTS, NOC, OSP, ADMIN, CUSTOMER SERVICE, FIELD, MARKETING",
		})
		return
	}

	// Validate role
	if !models.IsValidRole(req.Role) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid role. Valid roles: employee, hod, hr, ged, admin",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if user already exists
	var existingUser models.User
	err := config.UsersCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Email already registered",
		})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to hash password",
		})
		return
	}

	// Generate staff ID if not provided
	staffID := req.StaffID
	if staffID == "" {
		staffID, err = utils.GenerateStaffID(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to generate staff ID",
			})
			return
		}
	}

	// Set default values
	totalLeave := 28
	if req.TotalLeave > 0 {
		totalLeave = req.TotalLeave
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create user
	user := models.User{
		ID:         primitive.NewObjectID(),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Email:      req.Email,
		Password:   hashedPassword,
		StaffID:    staffID,
		Department: req.Department,
		Role:       req.Role,
		LeaveBalance: models.LeaveBalance{
			Total:     totalLeave,
			Available: totalLeave,
			Used:      0,
		},
		IsActive:  isActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err = config.UsersCollection.InsertOne(ctx, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create user",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "User created successfully",
		"user":    user.ToResponse(),
	})
}

// AdminUpdateUser updates user information (admin only)
func AdminUpdateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build update fields
	updateFields := bson.M{"updatedAt": time.Now()}

	if req.FirstName != "" {
		updateFields["firstName"] = req.FirstName
	}
	if req.LastName != "" {
		updateFields["lastName"] = req.LastName
	}
	if req.Email != "" {
		// Check if email is already taken by another user
		var existingUser models.User
		err := config.UsersCollection.FindOne(ctx, bson.M{
			"email": req.Email,
			"_id":   bson.M{"$ne": userID},
		}).Decode(&existingUser)
		if err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Email already exists for another user",
			})
			return
		}
		updateFields["email"] = req.Email
	}
	if req.Department != "" {
		if !models.IsValidDepartment(req.Department) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid department",
			})
			return
		}
		updateFields["department"] = req.Department
	}
	if req.Role != "" {
		if !models.IsValidRole(req.Role) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "Invalid role",
			})
			return
		}
		updateFields["role"] = req.Role
	}
	if req.StaffID != "" {
		updateFields["staffId"] = req.StaffID
	}
	if req.IsActive != nil {
		updateFields["isActive"] = *req.IsActive
	}

	// Update user
	result, err := config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update user",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	// Get updated user
	var user models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch updated user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User updated successfully",
		"user":    user.ToResponse(),
	})
}

// AdminDeactivateUser deactivates a user account (admin only)
func AdminDeactivateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Prevent admin from deactivating themselves
	currentUserID, err := middleware.GetCurrentUserID(c)
	if err == nil && currentUserID == userID {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Cannot deactivate your own account",
		})
		return
	}

	result, err := config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"isActive":  false,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to deactivate user",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User deactivated successfully",
	})
}

// AdminActivateUser reactivates a user account (admin only)
func AdminActivateUser(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"isActive":  true,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to activate user",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "User activated successfully",
	})
}

// AdminGetAllUsers gets all users including inactive ones (admin only)
func AdminGetAllUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Query parameters for filtering
	isActiveStr := c.Query("active")
	department := c.Query("department")
	role := c.Query("role")

	// Build filter
	filter := bson.M{}
	if isActiveStr != "" {
		if isActiveStr == "true" {
			filter["isActive"] = true
		} else if isActiveStr == "false" {
			filter["isActive"] = false
		}
	}
	if department != "" {
		filter["department"] = department
	}
	if role != "" {
		filter["role"] = role
	}

	cursor, err := config.UsersCollection.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch users",
		})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to decode users",
		})
		return
	}

	// Convert to response format
	userResponses := make([]models.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = user.ToResponse()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(userResponses),
		"users":   userResponses,
		"filters": gin.H{
			"active":     isActiveStr,
			"department": department,
			"role":       role,
		},
	})
}

// AdminUpdateUserLeaveBalance updates a user's leave balance (admin only)
func AdminUpdateUserLeaveBalance(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	var req AdminUpdateLeaveBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build update fields for leave balance
	updateFields := bson.M{"updatedAt": time.Now()}

	if req.Total != nil {
		updateFields["leaveBalance.total"] = *req.Total
	}
	if req.Available != nil {
		updateFields["leaveBalance.available"] = *req.Available
	}
	if req.Used != nil {
		updateFields["leaveBalance.used"] = *req.Used
	}

	// Update user leave balance
	result, err := config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update leave balance",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	// Get updated user
	var user models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch updated user",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Leave balance updated successfully",
		"user":    user.ToResponse(),
	})
}

// AdminResetUserPassword resets a user's password (admin only)
func AdminResetUserPassword(c *gin.Context) {
	idParam := c.Param("id")
	userID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid user ID",
		})
		return
	}

	var req struct {
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data. New password must be at least 6 characters.",
			"error":   err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Hash new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to hash password",
		})
		return
	}

	// Update password
	result, err := config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"password":  hashedPassword,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to reset password",
		})
		return
	}

	if result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Password reset successfully",
	})
}
