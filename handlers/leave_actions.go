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
)

// ApproveLeave approves a leave request
func ApproveLeave(c *gin.Context) {
	idParam := c.Param("id")
	leaveID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid leave ID",
		})
		return
	}

	var req models.ApproveRejectRequest
	c.ShouldBindJSON(&req)

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get leave
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Leave request not found",
		})
		return
	}

	// Determine approval role
	var approvalRole string
	switch leave.Stage {
	case 1:
		approvalRole = "HOD"
	case 2:
		approvalRole = "HR"
	case 3:
		approvalRole = "GED"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid approval stage",
		})
		return
	}

	// Add approval to flow
	approval := models.ApprovalStep{
		Approver: user.ID,
		Role:     approvalRole,
		Status:   "Approved",
		Comments: req.Comments,
		Date:     time.Now(),
	}

	leave.ApprovalFlow = append(leave.ApprovalFlow, approval)

	// Move to next stage or approve
	if leave.Stage < 3 {
		leave.Stage++
		leave.Status = "Pending"
		leave.IsEditable = leave.Stage < 2
	} else {
		leave.Status = "Approved"
		leave.IsEditable = false

		// Deduct from user's leave balance
		var employee models.User
		config.UsersCollection.FindOne(ctx, bson.M{"_id": leave.Employee}).Decode(&employee)

		_, err = config.UsersCollection.UpdateOne(
			ctx,
			bson.M{"_id": leave.Employee},
			bson.M{"$set": bson.M{
				"leaveBalance.available": employee.LeaveBalance.Available - leave.TotalDays,
				"leaveBalance.used":      employee.LeaveBalance.Used + leave.TotalDays,
				"updatedAt":              time.Now(),
			}},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to update leave balance",
			})
			return
		}
	}

	leave.UpdatedAt = time.Now()

	// Update leave
	_, err = config.LeavesCollection.UpdateOne(
		ctx,
		bson.M{"_id": leaveID},
		bson.M{"$set": bson.M{
			"status":       leave.Status,
			"stage":        leave.Stage,
			"approvalFlow": leave.ApprovalFlow,
			"isEditable":   leave.IsEditable,
			"updatedAt":    leave.UpdatedAt,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to approve leave",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Leave request approved by " + approvalRole,
		"leave": gin.H{
			"id":     leave.ID,
			"status": leave.Status,
			"stage":  leave.Stage,
		},
	})
}

// RejectLeave rejects a leave request
func RejectLeave(c *gin.Context) {
	idParam := c.Param("id")
	leaveID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid leave ID",
		})
		return
	}

	var req models.ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Comments == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Please provide reason for rejection",
		})
		return
	}

	user, err := middleware.GetCurrentUser(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not found",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get leave
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Leave request not found",
		})
		return
	}

	// Determine approval role
	var approvalRole string
	switch leave.Stage {
	case 1:
		approvalRole = "HOD"
	case 2:
		approvalRole = "HR"
	case 3:
		approvalRole = "GED"
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid approval stage",
		})
		return
	}

	// Add rejection to flow
	rejection := models.ApprovalStep{
		Approver: user.ID,
		Role:     approvalRole,
		Status:   "Rejected",
		Comments: req.Comments,
		Date:     time.Now(),
	}

	leave.ApprovalFlow = append(leave.ApprovalFlow, rejection)
	leave.Status = "Rejected"
	leave.UpdatedAt = time.Now()

	// Update leave
	_, err = config.LeavesCollection.UpdateOne(
		ctx,
		bson.M{"_id": leaveID},
		bson.M{"$set": bson.M{
			"status":       leave.Status,
			"approvalFlow": leave.ApprovalFlow,
			"updatedAt":    leave.UpdatedAt,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to reject leave",
		})
		return
	}

	// Refund leave days back to employee's balance
	_, err = config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": leave.Employee},
		bson.M{
			"$inc": bson.M{
				"leaveBalance.available": leave.TotalDays,
				"leaveBalance.used":      -leave.TotalDays,
			},
			"$set": bson.M{"updatedAt": time.Now()},
		},
	)
	if err != nil {
		// Log error but don't fail the rejection
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Leave request rejected by " + approvalRole + " (warning: balance refund may have failed)",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Leave request rejected by " + approvalRole + " and leave days refunded",
	})
}

// CancelLeave cancels a leave request
func CancelLeave(c *gin.Context) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get leave
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
			"message": "Not authorized to cancel this leave request",
		})
		return
	}

	// Check if can be cancelled
	if leave.Status == "Over" || leave.Status == "Rejected" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Cannot cancel completed or rejected leave",
		})
		return
	}

	// Refund days to balance for any cancellation (Pending, Approved, or Active)
	// Since we deduct immediately on creation, we need to refund on cancel
	if leave.Status == "Pending" || leave.Status == "Approved" || leave.Status == "Active" {
		_, err = config.UsersCollection.UpdateOne(
			ctx,
			bson.M{"_id": leave.Employee},
			bson.M{
				"$inc": bson.M{
					"leaveBalance.available": leave.TotalDays,
					"leaveBalance.used":      -leave.TotalDays,
				},
				"$set": bson.M{"updatedAt": time.Now()},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to refund leave balance",
			})
			return
		}
	}

	// Update leave status
	_, err = config.LeavesCollection.UpdateOne(
		ctx,
		bson.M{"_id": leaveID},
		bson.M{"$set": bson.M{
			"status":    "Cancelled",
			"updatedAt": time.Now(),
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to cancel leave",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Leave request cancelled successfully",
	})
}

// DeleteLeave deletes a leave request
func DeleteLeave(c *gin.Context) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get leave
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
			"message": "Not authorized to delete this leave request",
		})
		return
	}

	// Check if can be deleted
	if leave.Status == "Active" || leave.Status == "Approved" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Cannot delete approved or active leave. Please cancel instead.",
		})
		return
	}

	// Refund days if leave was pending (not rejected, as rejected already refunded)
	if leave.Status == "Pending" {
		_, err = config.UsersCollection.UpdateOne(
			ctx,
			bson.M{"_id": leave.Employee},
			bson.M{
				"$inc": bson.M{
					"leaveBalance.available": leave.TotalDays,
					"leaveBalance.used":      -leave.TotalDays,
				},
				"$set": bson.M{"updatedAt": time.Now()},
			},
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"message": "Failed to refund leave balance",
			})
			return
		}
	}

	// Delete leave
	_, err = config.LeavesCollection.DeleteOne(ctx, bson.M{"_id": leaveID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to delete leave",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Leave request deleted successfully",
	})
}
