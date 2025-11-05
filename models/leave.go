package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Leave represents a leave request
type Leave struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Employee       primitive.ObjectID `bson:"employee" json:"employee"`
	LeaveType      string             `bson:"leaveType" json:"leaveType" binding:"required"`
	OtherLeaveType string             `bson:"otherLeaveType,omitempty" json:"otherLeaveType,omitempty"`
	FromDate       time.Time          `bson:"fromDate" json:"fromDate" binding:"required"`
	ToDate         time.Time          `bson:"toDate" json:"toDate" binding:"required"`
	TotalDays      int                `bson:"totalDays" json:"totalDays"`
	Reason         string             `bson:"reason" json:"reason" binding:"required"`
	Reliever       primitive.ObjectID `bson:"reliever" json:"reliever" binding:"required"`
	Status         string             `bson:"status" json:"status"`
	Stage          int                `bson:"stage" json:"stage"`
	ApprovalFlow   []ApprovalStep     `bson:"approvalFlow" json:"approvalFlow"`

	// Multi-stage approval tracking
	HODApprovalStatus  string             `bson:"hodApprovalStatus" json:"hodApprovalStatus"`
	HODApprovalDate    *time.Time         `bson:"hodApprovalDate,omitempty" json:"hodApprovalDate,omitempty"`
	HODApprovalComment string             `bson:"hodApprovalComment,omitempty" json:"hodApprovalComment,omitempty"`
	HODApprover        primitive.ObjectID `bson:"hodApprover,omitempty" json:"hodApprover,omitempty"`

	HRApprovalStatus  string             `bson:"hrApprovalStatus" json:"hrApprovalStatus"`
	HRApprovalDate    *time.Time         `bson:"hrApprovalDate,omitempty" json:"hrApprovalDate,omitempty"`
	HRApprovalComment string             `bson:"hrApprovalComment,omitempty" json:"hrApprovalComment,omitempty"`
	HRApprover        primitive.ObjectID `bson:"hrApprover,omitempty" json:"hrApprover,omitempty"`

	GEDApprovalStatus  string             `bson:"gedApprovalStatus" json:"gedApprovalStatus"`
	GEDApprovalDate    *time.Time         `bson:"gedApprovalDate,omitempty" json:"gedApprovalDate,omitempty"`
	GEDApprovalComment string             `bson:"gedApprovalComment,omitempty" json:"gedApprovalComment,omitempty"`
	GEDApprover        primitive.ObjectID `bson:"gedApprover,omitempty" json:"gedApprover,omitempty"`

	IsEditable bool      `bson:"isEditable" json:"isEditable"`
	IsActive   bool      `bson:"isActive" json:"isActive"`
	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt" json:"updatedAt"`
}

// ApprovalStep represents one approval in the workflow
type ApprovalStep struct {
	Approver primitive.ObjectID `bson:"approver" json:"approver"`
	Role     string             `bson:"role" json:"role"`
	Status   string             `bson:"status" json:"status"`
	Comments string             `bson:"comments,omitempty" json:"comments,omitempty"`
	Date     time.Time          `bson:"date" json:"date"`
}

// LeaveResponse includes populated user data
type LeaveResponse struct {
	ID             primitive.ObjectID   `json:"id"`
	Employee       UserResponse         `json:"employee"`
	LeaveType      string               `json:"leaveType"`
	OtherLeaveType string               `json:"otherLeaveType,omitempty"`
	FromDate       time.Time            `json:"fromDate"`
	ToDate         time.Time            `json:"toDate"`
	TotalDays      int                  `json:"totalDays"`
	Reason         string               `json:"reason"`
	Reliever       UserResponse         `json:"reliever"`
	Status         string               `json:"status"`
	Stage          int                  `json:"stage"`
	ApprovalFlow   []ApprovalStepDetail `json:"approvalFlow"`

	// Multi-stage approval tracking
	HODApprovalStatus  string        `json:"hodApprovalStatus"`
	HODApprovalDate    *time.Time    `json:"hodApprovalDate,omitempty"`
	HODApprovalComment string        `json:"hodApprovalComment,omitempty"`
	HODApprover        *ApproverInfo `json:"hodApprover,omitempty"`

	HRApprovalStatus  string        `json:"hrApprovalStatus"`
	HRApprovalDate    *time.Time    `json:"hrApprovalDate,omitempty"`
	HRApprovalComment string        `json:"hrApprovalComment,omitempty"`
	HRApprover        *ApproverInfo `json:"hrApprover,omitempty"`

	GEDApprovalStatus  string        `json:"gedApprovalStatus"`
	GEDApprovalDate    *time.Time    `json:"gedApprovalDate,omitempty"`
	GEDApprovalComment string        `json:"gedApprovalComment,omitempty"`
	GEDApprover        *ApproverInfo `json:"gedApprover,omitempty"`

	IsEditable bool      `json:"isEditable"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// ApprovalStepDetail includes approver details
type ApprovalStepDetail struct {
	Approver ApproverInfo `json:"approver"`
	Role     string       `json:"role"`
	Status   string       `json:"status"`
	Comments string       `json:"comments,omitempty"`
	Date     time.Time    `json:"date"`
}

// ApproverInfo represents minimal approver information
type ApproverInfo struct {
	ID        primitive.ObjectID `json:"id"`
	FirstName string             `json:"firstName"`
	LastName  string             `json:"lastName"`
	Role      string             `json:"role"`
}

// CreateLeaveRequest represents leave creation data
type CreateLeaveRequest struct {
	LeaveType      string `json:"leaveType" binding:"required"`
	OtherLeaveType string `json:"otherLeaveType,omitempty"`
	FromDate       string `json:"fromDate" binding:"required"`
	ToDate         string `json:"toDate" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	Reliever       string `json:"reliever" binding:"required"`
}

// UpdateLeaveRequest represents leave update data
type UpdateLeaveRequest struct {
	LeaveType      string `json:"leaveType"`
	OtherLeaveType string `json:"otherLeaveType"`
	FromDate       string `json:"fromDate"`
	ToDate         string `json:"toDate"`
	Reason         string `json:"reason"`
	Reliever       string `json:"reliever"`
}

// ApproveRejectRequest represents approval/rejection data
type ApproveRejectRequest struct {
	Comments string `json:"comments"`
}

// Valid leave types
var ValidLeaveTypes = []string{
	"Annual Leave", "Sick Leave", "Casual Leave", "Other",
}

// Valid leave statuses
var ValidLeaveStatuses = []string{
	"Pending", "Active", "Approved", "Rejected", "Over", "Cancelled",
}

// Valid approval roles
var ValidApprovalRoles = []string{
	"HOD", "HR", "GED",
}

// Valid approval stage statuses
var ValidApprovalStageStatuses = []string{
	"pending", "approved", "rejected",
}

// IsValidLeaveType checks if leave type is valid
func IsValidLeaveType(leaveType string) bool {
	for _, lt := range ValidLeaveTypes {
		if lt == leaveType {
			return true
		}
	}
	return false
}

// CalculateDays calculates total days between two dates excluding weekends
func CalculateDays(from, to time.Time) int {
	if to.Before(from) {
		return 0
	}

	workdays := 0
	current := from

	// Iterate through each day and count only weekdays
	for !current.After(to) {
		// Check if it's a weekday (Monday = 1, Sunday = 0)
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			workdays++
		}
		// Move to next day
		current = current.Add(24 * time.Hour)
	}

	if workdays < 1 {
		return 0
	}
	return workdays
}
