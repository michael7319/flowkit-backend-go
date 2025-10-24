package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/middleware"
	"github.com/flowkit/backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetUsers gets all active users
func GetUsers(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := config.UsersCollection.Find(
		ctx,
		bson.M{"isActive": true},
		options.Find().SetSort(bson.D{{Key: "firstName", Value: 1}}),
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
	})
}

// GetRelievers gets potential relievers (all users except current)
func GetRelievers(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := config.UsersCollection.Find(
		ctx,
		bson.M{
			"_id":      bson.M{"$ne": userID},
			"isActive": true,
		},
		options.Find().SetSort(bson.D{{Key: "firstName", Value: 1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch relievers",
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

	// Convert to response format (minimal info for relievers)
	relievers := make([]gin.H, len(users))
	for i, user := range users {
		relievers[i] = gin.H{
			"id":         user.ID,
			"firstName":  user.FirstName,
			"lastName":   user.LastName,
			"email":      user.Email,
			"department": user.Department,
			"staffId":    user.StaffID,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"count":     len(relievers),
		"relievers": relievers,
	})
}

// GetUserByID gets user by ID
func GetUserByID(c *gin.Context) {
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

	var user models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user":    user.ToResponse(),
	})
}

// UpdateProfile updates user profile
func UpdateProfile(c *gin.Context) {
	var req models.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
		})
		return
	}

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found",
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
		updateFields["email"] = req.Email
	}
	if req.StaffID != "" {
		updateFields["staffId"] = req.StaffID
	}

	// Update user
	_, err = config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update profile",
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
		"message": "Profile updated successfully",
		"user":    user.ToResponse(),
	})
}

// UploadSignature uploads user signature
func UploadSignature(c *gin.Context) {
	var req models.UploadSignatureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
		})
		return
	}

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Update signature
	_, err = config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{"$set": bson.M{
			"signature": req.Signature,
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to upload signature",
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
		"message": "Signature uploaded successfully",
		"user":    user.ToResponse(),
	})
}
