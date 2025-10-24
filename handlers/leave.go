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

// CreateLeave creates a new leave request
func CreateLeave(c *gin.Context) {
	var req models.CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request data",
			"error":   err.Error(),
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

	// Validate leave type
	if !models.IsValidLeaveType(req.LeaveType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid leave type",
		})
		return
	}

	// Parse dates
	fromDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid from date format. Use YYYY-MM-DD",
		})
		return
	}

	toDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid to date format. Use YYYY-MM-DD",
		})
		return
	}

	// Validate dates
	now := time.Now().Truncate(24 * time.Hour)
	if fromDate.Before(now) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Start date cannot be in the past",
		})
		return
	}

	if toDate.Before(fromDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "End date must be after start date",
		})
		return
	}

	// Calculate total days
	totalDays := models.CalculateDays(fromDate, toDate)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check user leave balance
	var user models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch user",
		})
		return
	}

	if user.LeaveBalance.Available < totalDays {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Insufficient leave balance. You have " + string(rune(user.LeaveBalance.Available)) + " days available.",
		})
		return
	}

	// Validate reliever
	relieverID, err := primitive.ObjectIDFromHex(req.Reliever)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid reliever ID",
		})
		return
	}

	var reliever models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": relieverID}).Decode(&reliever)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid reliever selected",
		})
		return
	}

	// Create leave
	leave := models.Leave{
		ID:             primitive.NewObjectID(),
		Employee:       userID,
		LeaveType:      req.LeaveType,
		OtherLeaveType: req.OtherLeaveType,
		FromDate:       fromDate,
		ToDate:         toDate,
		TotalDays:      totalDays,
		Reason:         req.Reason,
		Reliever:       relieverID,
		Status:         "Pending",
		Stage:          1, // Start at HOD approval
		ApprovalFlow:   []models.ApprovalStep{},
		IsEditable:     true,
		IsActive:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	_, err = config.LeavesCollection.InsertOne(ctx, leave)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to create leave request",
		})
		return
	}

	// Deduct leave days from available balance immediately
	_, err = config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": userID},
		bson.M{
			"$inc": bson.M{
				"leaveBalance.available": -totalDays,
				"leaveBalance.used":      totalDays,
			},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		// Rollback: delete the leave request if balance update fails
		config.LeavesCollection.DeleteOne(ctx, bson.M{"_id": leave.ID})
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update leave balance",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Leave request created successfully",
		"leave": gin.H{
			"id":         leave.ID,
			"leaveType":  leave.LeaveType,
			"fromDate":   leave.FromDate,
			"toDate":     leave.ToDate,
			"totalDays":  leave.TotalDays,
			"status":     leave.Status,
			"stage":      leave.Stage,
			"isEditable": leave.IsEditable,
		},
	})
}

// GetMyLeaves gets current user's leave requests
func GetMyLeaves(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := config.LeavesCollection.Find(
		ctx,
		bson.M{"employee": userID},
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch leaves",
		})
		return
	}
	defer cursor.Close(ctx)

	var leaves []models.Leave
	if err := cursor.All(ctx, &leaves); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to decode leaves",
		})
		return
	}

	// Update status for active leaves and populate user data
	leaveResponses := make([]gin.H, len(leaves))
	for i, leave := range leaves {
		// Check and update status
		updateLeaveStatus(&leave)
		config.LeavesCollection.UpdateOne(ctx, bson.M{"_id": leave.ID}, bson.M{"$set": bson.M{
			"status":   leave.Status,
			"isActive": leave.IsActive,
		}})

		// Get employee info
		var employee models.User
		config.UsersCollection.FindOne(ctx, bson.M{"_id": leave.Employee}).Decode(&employee)

		// Get reliever info
		var reliever models.User
		config.UsersCollection.FindOne(ctx, bson.M{"_id": leave.Reliever}).Decode(&reliever)

		leaveResponses[i] = gin.H{
			"id": leave.ID,
			"employee": gin.H{
				"id":         employee.ID,
				"firstName":  employee.FirstName,
				"lastName":   employee.LastName,
				"department": employee.Department,
			},
			"leaveType":      leave.LeaveType,
			"otherLeaveType": leave.OtherLeaveType,
			"fromDate":       leave.FromDate,
			"toDate":         leave.ToDate,
			"totalDays":      leave.TotalDays,
			"reason":         leave.Reason,
			"reliever": gin.H{
				"id":        reliever.ID,
				"firstName": reliever.FirstName,
				"lastName":  reliever.LastName,
			},
			"status":       leave.Status,
			"stage":        leave.Stage,
			"approvalFlow": leave.ApprovalFlow,
			"isEditable":   leave.IsEditable,
			"isActive":     leave.IsActive,
			"createdAt":    leave.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(leaveResponses),
		"leaves":  leaveResponses,
	})
}

// updateLeaveStatus checks and updates leave status
func updateLeaveStatus(leave *models.Leave) {
	now := time.Now()
	start := leave.FromDate
	end := leave.ToDate

	if leave.Status == "Approved" && !now.Before(start) && !now.After(end) {
		leave.Status = "Active"
		leave.IsActive = true
	} else if leave.Status == "Active" && now.After(end) {
		leave.Status = "Over"
		leave.IsActive = false
	}
}

// GetAllLeaves gets all leave requests (for approvers)
func GetAllLeaves(c *gin.Context) {
	status := c.Query("status")
	department := c.Query("department")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{}
	if status != "" {
		filter["status"] = status
	}

	cursor, err := config.LeavesCollection.Find(
		ctx,
		filter,
		options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}}),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch leaves",
		})
		return
	}
	defer cursor.Close(ctx)

	var leaves []models.Leave
	if err := cursor.All(ctx, &leaves); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to decode leaves",
		})
		return
	}

	// Populate user data and filter by department if needed
	leaveResponses := []gin.H{}
	for _, leave := range leaves {
		var employee models.User
		config.UsersCollection.FindOne(ctx, bson.M{"_id": leave.Employee}).Decode(&employee)

		// Filter by department if specified
		if department != "" && employee.Department != department {
			continue
		}

		var reliever models.User
		config.UsersCollection.FindOne(ctx, bson.M{"_id": leave.Reliever}).Decode(&reliever)

		leaveResponses = append(leaveResponses, gin.H{
			"id": leave.ID,
			"employee": gin.H{
				"id":         employee.ID,
				"firstName":  employee.FirstName,
				"lastName":   employee.LastName,
				"department": employee.Department,
			},
			"leaveType": leave.LeaveType,
			"fromDate":  leave.FromDate,
			"toDate":    leave.ToDate,
			"totalDays": leave.TotalDays,
			"reliever": gin.H{
				"firstName": reliever.FirstName,
				"lastName":  reliever.LastName,
			},
			"status":     leave.Status,
			"stage":      leave.Stage,
			"isEditable": leave.IsEditable,
			"createdAt":  leave.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"count":   len(leaveResponses),
		"leaves":  leaveResponses,
	})
}

// UpdateLeave updates a leave request (only editable before Stage 2)
func UpdateLeave(c *gin.Context) {
	idParam := c.Param("id")
	leaveID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid leave ID",
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

	var req models.CreateLeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid input data",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get existing leave
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Leave request not found",
		})
		return
	}

	// Check ownership
	user, _ := middleware.GetCurrentUser(c)
	if leave.Employee != userID && user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "Not authorized to update this leave request",
		})
		return
	}

	// Check if editable
	if !leave.IsEditable || leave.Stage >= 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Leave request cannot be edited at this stage",
		})
		return
	}

	// Validate dates
	startDate, err := time.Parse("2006-01-02", req.FromDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid start date format",
		})
		return
	}

	endDate, err := time.Parse("2006-01-02", req.ToDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid end date format",
		})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "End date cannot be before start date",
		})
		return
	}

	// Calculate total days
	totalDays := models.CalculateDays(startDate, endDate)

	// Get user to check balance
	var employee models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch user details",
		})
		return
	}

	// Calculate days difference (positive = need more days, negative = refund days)
	daysDifference := totalDays - leave.TotalDays

	// Check if sufficient leave balance if days increased
	if daysDifference > 0 && employee.LeaveBalance.Available < daysDifference {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Insufficient leave balance. Need " + string(rune(daysDifference)) + " more days.",
		})
		return
	}

	// Update leave fields
	relieverID, err := primitive.ObjectIDFromHex(req.Reliever)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid reliever ID",
		})
		return
	}

	update := bson.M{
		"leaveType":      req.LeaveType,
		"otherLeaveType": req.OtherLeaveType,
		"fromDate":       startDate,
		"toDate":         endDate,
		"totalDays":      totalDays,
		"reliever":       relieverID,
		"reason":         req.Reason,
		"updatedAt":      time.Now(),
	}

	_, err = config.LeavesCollection.UpdateOne(
		ctx,
		bson.M{"_id": leaveID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to update leave request",
		})
		return
	}

	// Adjust user's leave balance if days changed
	if daysDifference != 0 {
		_, err = config.UsersCollection.UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{
				"$inc": bson.M{
					"leaveBalance.available": -daysDifference, // Negative if increasing days, positive if reducing
					"leaveBalance.used":      daysDifference,  // Positive if increasing days, negative if reducing
				},
				"$set": bson.M{"updatedAt": time.Now()},
			},
		)
		if err != nil {
			// Log error but don't fail the update
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Leave request updated but balance adjustment may have failed",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Leave request updated successfully",
	})
}
