package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// HODApproveLeave allows HOD to approve leave requests from their department
func HODApproveLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("userID")

	// Validate leave ID
	leaveObjID, err := primitive.ObjectIDFromHex(leaveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave ID"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get comment from request
	var req models.ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Comments are optional
		req.Comments = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the current user (HOD)
	var hod models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userObjID}).Decode(&hod)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is actually a HOD
	if !hod.IsHOD {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only HODs can approve at this stage"})
		return
	}

	// Get the leave request
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveObjID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Get employee details to check department
	var employee models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": leave.Employee}).Decode(&employee)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Check if HOD is from the same department as the employee
	if hod.Department != employee.Department {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only approve leave requests from your department"})
		return
	}

	// Check if already processed
	if leave.HODApprovalStatus == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already approved by HOD"})
		return
	}
	if leave.HODApprovalStatus == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already rejected by HOD"})
		return
	}

	// Update leave with HOD approval
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"hodApprovalStatus":  "approved",
			"hodApprovalDate":    now,
			"hodApprovalComment": req.Comments,
			"hodApprover":        userObjID,
			"updatedAt":          now,
		},
	}

	_, err = config.LeavesCollection.UpdateOne(ctx, bson.M{"_id": leaveObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request approved by HOD successfully",
		"leaveId": leaveID,
	})
}

// HODRejectLeave allows HOD to reject leave requests from their department
func HODRejectLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("userID")

	// Validate leave ID
	leaveObjID, err := primitive.ObjectIDFromHex(leaveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave ID"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get comment from request
	var req models.ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Comments are optional but recommended for rejections
		req.Comments = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the current user (HOD)
	var hod models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userObjID}).Decode(&hod)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Check if user is actually a HOD
	if !hod.IsHOD {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only HODs can reject at this stage"})
		return
	}

	// Get the leave request
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveObjID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Get employee details to check department
	var employee models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": leave.Employee}).Decode(&employee)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		return
	}

	// Check if HOD is from the same department as the employee
	if hod.Department != employee.Department {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only reject leave requests from your department"})
		return
	}

	// Check if already processed
	if leave.HODApprovalStatus == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already rejected by HOD"})
		return
	}

	// Refund leave days to employee
	_, err = config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": leave.Employee},
		bson.M{
			"$inc": bson.M{
				"leaveBalance.available": leave.TotalDays,
				"leaveBalance.used":      -leave.TotalDays,
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee leave balance"})
		return
	}

	// Update leave with HOD rejection
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"hodApprovalStatus":  "rejected",
			"hodApprovalDate":    now,
			"hodApprovalComment": req.Comments,
			"hodApprover":        userObjID,
			"status":             "Rejected",
			"updatedAt":          now,
		},
	}

	_, err = config.LeavesCollection.UpdateOne(ctx, bson.M{"_id": leaveObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request rejected by HOD successfully",
		"leaveId": leaveID,
	})
}

// HRApproveLeave allows HR to approve leave requests that have been approved by HOD
func HRApproveLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("userID")

	// Validate leave ID
	leaveObjID, err := primitive.ObjectIDFromHex(leaveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave ID"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get comment from request
	var req models.ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comments = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the leave request
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveObjID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Check if HOD has approved first
	if leave.HODApprovalStatus != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Leave request must be approved by HOD first"})
		return
	}

	// Check if already processed by HR
	if leave.HRApprovalStatus == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already approved by HR"})
		return
	}
	if leave.HRApprovalStatus == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already rejected by HR"})
		return
	}

	// Update leave with HR approval
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"hrApprovalStatus":  "approved",
			"hrApprovalDate":    now,
			"hrApprovalComment": req.Comments,
			"hrApprover":        userObjID,
			"updatedAt":         now,
		},
	}

	_, err = config.LeavesCollection.UpdateOne(ctx, bson.M{"_id": leaveObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request approved by HR successfully",
		"leaveId": leaveID,
	})
}

// HRRejectLeave allows HR to reject leave requests
func HRRejectLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("userID")

	// Validate leave ID
	leaveObjID, err := primitive.ObjectIDFromHex(leaveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave ID"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get comment from request
	var req models.ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comments = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the leave request
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveObjID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Check if HOD has approved first
	if leave.HODApprovalStatus != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Leave request must be approved by HOD before HR can review"})
		return
	}

	// Check if already processed by HR
	if leave.HRApprovalStatus == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already rejected by HR"})
		return
	}

	// Refund leave days to employee
	_, err = config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": leave.Employee},
		bson.M{
			"$inc": bson.M{
				"leaveBalance.available": leave.TotalDays,
				"leaveBalance.used":      -leave.TotalDays,
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee leave balance"})
		return
	}

	// Update leave with HR rejection
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"hrApprovalStatus":  "rejected",
			"hrApprovalDate":    now,
			"hrApprovalComment": req.Comments,
			"hrApprover":        userObjID,
			"status":            "Rejected",
			"updatedAt":         now,
		},
	}

	_, err = config.LeavesCollection.UpdateOne(ctx, bson.M{"_id": leaveObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request rejected by HR successfully",
		"leaveId": leaveID,
	})
}

// GEDApproveLeave allows GED to give final approval to leave requests
func GEDApproveLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("userID")

	// Validate leave ID
	leaveObjID, err := primitive.ObjectIDFromHex(leaveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave ID"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get comment from request
	var req models.ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comments = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the leave request
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveObjID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Check if HOD and HR have approved first
	if leave.HODApprovalStatus != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Leave request must be approved by HOD first"})
		return
	}
	if leave.HRApprovalStatus != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Leave request must be approved by HR first"})
		return
	}

	// Check if already processed by GED
	if leave.GEDApprovalStatus == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already approved by GED"})
		return
	}
	if leave.GEDApprovalStatus == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already rejected by GED"})
		return
	}

	// Update leave with GED approval - this makes it fully approved
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"gedApprovalStatus":  "approved",
			"gedApprovalDate":    now,
			"gedApprovalComment": req.Comments,
			"gedApprover":        userObjID,
			"status":             "Approved", // Only now is it truly approved
			"updatedAt":          now,
		},
	}

	_, err = config.LeavesCollection.UpdateOne(ctx, bson.M{"_id": leaveObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request approved by GED successfully - Leave is now fully approved",
		"leaveId": leaveID,
	})
}

// GEDRejectLeave allows GED to reject leave requests
func GEDRejectLeave(c *gin.Context) {
	leaveID := c.Param("id")
	userID := c.GetString("userID")

	// Validate leave ID
	leaveObjID, err := primitive.ObjectIDFromHex(leaveID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid leave ID"})
		return
	}

	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get comment from request
	var req models.ApproveRejectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comments = ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get the leave request
	var leave models.Leave
	err = config.LeavesCollection.FindOne(ctx, bson.M{"_id": leaveObjID}).Decode(&leave)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// Check if HOD and HR have approved first
	if leave.HODApprovalStatus != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Leave request must be approved by HOD before GED can review"})
		return
	}
	if leave.HRApprovalStatus != "approved" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Leave request must be approved by HR before GED can review"})
		return
	}

	// Check if already processed by GED
	if leave.GEDApprovalStatus == "rejected" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave request already rejected by GED"})
		return
	}

	// Refund leave days to employee
	_, err = config.UsersCollection.UpdateOne(
		ctx,
		bson.M{"_id": leave.Employee},
		bson.M{
			"$inc": bson.M{
				"leaveBalance.available": leave.TotalDays,
				"leaveBalance.used":      -leave.TotalDays,
			},
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update employee leave balance"})
		return
	}

	// Update leave with GED rejection
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"gedApprovalStatus":  "rejected",
			"gedApprovalDate":    now,
			"gedApprovalComment": req.Comments,
			"gedApprover":        userObjID,
			"status":             "Rejected",
			"updatedAt":          now,
		},
	}

	_, err = config.LeavesCollection.UpdateOne(ctx, bson.M{"_id": leaveObjID}, update)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request rejected by GED successfully",
		"leaveId": leaveID,
	})
}

// GetHODLeaves returns leave requests for HOD to review (from their department)
func GetHODLeaves(c *gin.Context) {
	userID := c.GetString("userID")
	userObjID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get HOD details
	var hod models.User
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userObjID}).Decode(&hod)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !hod.IsHOD {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only HODs can access this endpoint"})
		return
	}

	// Get all employees from the HOD's department
	cursor, err := config.UsersCollection.Find(ctx, bson.M{
		"department": hod.Department,
		"isActive":   true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch department employees"})
		return
	}
	defer cursor.Close(ctx)

	var departmentEmployeeIDs []primitive.ObjectID
	for cursor.Next(ctx) {
		var user models.User
		if err := cursor.Decode(&user); err == nil {
			departmentEmployeeIDs = append(departmentEmployeeIDs, user.ID)
		}
	}

	// Get leave requests from department employees
	leaveCursor, err := config.LeavesCollection.Find(ctx, bson.M{
		"employee": bson.M{"$in": departmentEmployeeIDs},
		"isActive": true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leave requests"})
		return
	}
	defer leaveCursor.Close(ctx)

	var leaves []models.Leave
	if err := leaveCursor.All(ctx, &leaves); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode leave requests"})
		return
	}

	// Populate employee data for each leave
	leaveResponses := make([]gin.H, len(leaves))
	for i, leave := range leaves {
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
			"status":             leave.Status,
			"hodApprovalStatus":  leave.HODApprovalStatus,
			"hodApprovalDate":    leave.HODApprovalDate,
			"hodApprovalComment": leave.HODApprovalComment,
			"hrApprovalStatus":   leave.HRApprovalStatus,
			"hrApprovalDate":     leave.HRApprovalDate,
			"hrApprovalComment":  leave.HRApprovalComment,
			"gedApprovalStatus":  leave.GEDApprovalStatus,
			"gedApprovalDate":    leave.GEDApprovalDate,
			"gedApprovalComment": leave.GEDApprovalComment,
			"isEditable":         leave.IsEditable,
			"isActive":           leave.IsActive,
			"createdAt":          leave.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"leaves":  leaveResponses,
	})
}

// GetHRLeaves returns leave requests for HR to review (HOD approved only)
func GetHRLeaves(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get leave requests that have been approved by HOD but not yet processed by HR
	cursor, err := config.LeavesCollection.Find(ctx, bson.M{
		"hodApprovalStatus": "approved",
		"isActive":          true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leave requests"})
		return
	}
	defer cursor.Close(ctx)

	var leaves []models.Leave
	if err := cursor.All(ctx, &leaves); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode leave requests"})
		return
	}

	// Populate employee data for each leave
	leaveResponses := make([]gin.H, len(leaves))
	for i, leave := range leaves {
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
			"status":             leave.Status,
			"hodApprovalStatus":  leave.HODApprovalStatus,
			"hodApprovalDate":    leave.HODApprovalDate,
			"hodApprovalComment": leave.HODApprovalComment,
			"hrApprovalStatus":   leave.HRApprovalStatus,
			"hrApprovalDate":     leave.HRApprovalDate,
			"hrApprovalComment":  leave.HRApprovalComment,
			"gedApprovalStatus":  leave.GEDApprovalStatus,
			"gedApprovalDate":    leave.GEDApprovalDate,
			"gedApprovalComment": leave.GEDApprovalComment,
			"isEditable":         leave.IsEditable,
			"isActive":           leave.IsActive,
			"createdAt":          leave.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"leaves":  leaveResponses,
	})
}

// GetGEDLeaves returns leave requests for GED to review (HOD and HR approved only)
func GetGEDLeaves(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get leave requests that have been approved by both HOD and HR
	cursor, err := config.LeavesCollection.Find(ctx, bson.M{
		"hodApprovalStatus": "approved",
		"hrApprovalStatus":  "approved",
		"isActive":          true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leave requests"})
		return
	}
	defer cursor.Close(ctx)

	var leaves []models.Leave
	if err := cursor.All(ctx, &leaves); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode leave requests"})
		return
	}

	// Populate employee data for each leave
	leaveResponses := make([]gin.H, len(leaves))
	for i, leave := range leaves {
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
			"status":             leave.Status,
			"hodApprovalStatus":  leave.HODApprovalStatus,
			"hodApprovalDate":    leave.HODApprovalDate,
			"hodApprovalComment": leave.HODApprovalComment,
			"hrApprovalStatus":   leave.HRApprovalStatus,
			"hrApprovalDate":     leave.HRApprovalDate,
			"hrApprovalComment":  leave.HRApprovalComment,
			"gedApprovalStatus":  leave.GEDApprovalStatus,
			"gedApprovalDate":    leave.GEDApprovalDate,
			"gedApprovalComment": leave.GEDApprovalComment,
			"isEditable":         leave.IsEditable,
			"isActive":           leave.IsActive,
			"createdAt":          leave.CreatedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"leaves":  leaveResponses,
	})
}
