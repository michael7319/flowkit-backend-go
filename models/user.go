package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FirstName    string             `bson:"firstName" json:"firstName" binding:"required"`
	LastName     string             `bson:"lastName" json:"lastName" binding:"required"`
	Email        string             `bson:"email" json:"email" binding:"required,email"`
	Password     string             `bson:"password" json:"-"`
	StaffID      string             `bson:"staffId" json:"staffId"`
	Department   string             `bson:"department" json:"department" binding:"required"`
	Role         string             `bson:"role" json:"role"`
	Signature    string             `bson:"signature,omitempty" json:"signature,omitempty"`
	LeaveBalance LeaveBalance       `bson:"leaveBalance" json:"leaveBalance"`
	IsActive     bool               `bson:"isActive" json:"isActive"`
	CreatedAt    time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time          `bson:"updatedAt" json:"updatedAt"`
}

// LeaveBalance represents user's leave balance
type LeaveBalance struct {
	Total     int `bson:"total" json:"total"`
	Available int `bson:"available" json:"available"`
	Used      int `bson:"used" json:"used"`
}

// UserResponse is the response structure (without password)
type UserResponse struct {
	ID           primitive.ObjectID `json:"id"`
	FirstName    string             `json:"firstName"`
	LastName     string             `json:"lastName"`
	Email        string             `json:"email"`
	StaffID      string             `json:"staffId"`
	Department   string             `json:"department"`
	Role         string             `json:"role"`
	Signature    string             `json:"signature,omitempty"`
	LeaveBalance LeaveBalance       `json:"leaveBalance"`
	IsActive     bool               `json:"isActive"`
	CreatedAt    time.Time          `json:"createdAt"`
	UpdatedAt    time.Time          `json:"updatedAt"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Email:        u.Email,
		StaffID:      u.StaffID,
		Department:   u.Department,
		Role:         u.Role,
		Signature:    u.Signature,
		LeaveBalance: u.LeaveBalance,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	FirstName  string `json:"firstName" binding:"required"`
	LastName   string `json:"lastName" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=6"`
	Department string `json:"department" binding:"required"`
}

// LoginRequest represents login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents profile update data
type UpdateProfileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	StaffID   string `json:"staffId"`
}

// UpdatePasswordRequest represents password change data
type UpdatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" binding:"required"`
	NewPassword     string `json:"newPassword" binding:"required,min=6"`
}

// UploadSignatureRequest represents signature upload data
type UploadSignatureRequest struct {
	Signature string `json:"signature" binding:"required"`
}

// Valid departments
var ValidDepartments = []string{
	"VAS", "VOICE", "ACCOUNTS", "NOC", "OSP",
	"ADMIN", "CUSTOMER SERVICE", "FIELD", "MARKETING",
}

// Valid roles
var ValidRoles = []string{
	"employee", "hod", "hr", "ged", "admin",
}

// IsValidDepartment checks if department is valid
func IsValidDepartment(dept string) bool {
	for _, d := range ValidDepartments {
		if d == dept {
			return true
		}
	}
	return false
}

// IsValidRole checks if role is valid
func IsValidRole(role string) bool {
	for _, r := range ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}
