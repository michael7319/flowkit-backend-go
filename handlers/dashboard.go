package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/flowkit/backend/config"
	"github.com/flowkit/backend/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetDashboardStats returns dashboard statistics for the user
func GetDashboardStats(c *gin.Context) {
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

	// Get user's leave balance
	var user bson.M
	err = config.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch user data",
		})
		return
	}

	// Extract leave balance
	totalLeaves := 0
	availableLeaves := 0
	usedLeaves := 0
	if leaveBalance, ok := user["leaveBalance"].(bson.M); ok {
		if total, ok := leaveBalance["total"].(int32); ok {
			totalLeaves = int(total)
		}
		if available, ok := leaveBalance["available"].(int32); ok {
			availableLeaves = int(available)
		}
		if used, ok := leaveBalance["used"].(int32); ok {
			usedLeaves = int(used)
		}
	}

	// Approved leaves count
	approvedCount, _ := config.LeavesCollection.CountDocuments(ctx, bson.M{
		"employee": userID,
		"status":   "Approved",
	})

	// Rejected leaves count
	rejectedCount, _ := config.LeavesCollection.CountDocuments(ctx, bson.M{
		"employee": userID,
		"status":   "Rejected",
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"totalLeaves":     totalLeaves,
			"availableLeaves": availableLeaves,
			"usedLeaves":      usedLeaves,
			"approvedLeaves":  approvedCount,
			"rejectedLeaves":  rejectedCount,
		},
	})
}

// GetLeaveProgress returns current leave approval progress
func GetLeaveProgress(c *gin.Context) {
	idParam := c.Param("id")
	leaveID, err := primitive.ObjectIDFromHex(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid leave ID",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get leave with populated approval flow
	pipeline := []bson.M{
		{"$match": bson.M{"_id": leaveID}},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "approvalFlow.approver",
				"foreignField": "_id",
				"as":           "approvers",
			},
		},
	}

	cursor, err := config.LeavesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch leave progress",
		})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil || len(results) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "Leave not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results[0],
	})
}

// GetGraphData returns leave statistics for graph visualization
func GetGraphData(c *gin.Context) {
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

	// Get current year
	year := time.Now().Year()

	// Aggregate leaves by month - get total days taken per month per status
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"employee": userID,
				"$expr": bson.M{
					"$eq": bson.A{
						bson.M{"$year": "$fromDate"},
						year,
					},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"month":  bson.M{"$month": "$fromDate"},
					"status": "$status",
				},
				"count":     bson.M{"$sum": 1},
				"totalDays": bson.M{"$sum": "$totalDays"}, // Sum up actual days taken
			},
		},
		{
			"$sort": bson.M{"_id.month": 1},
		},
	}

	cursor, err := config.LeavesCollection.Aggregate(ctx, pipeline)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to fetch graph data",
		})
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M
	cursor.All(ctx, &results)

	// Format data for graph - track days taken per month by status
	monthlyData := make(map[string]map[string]int)
	months := []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

	for _, month := range months {
		monthlyData[month] = map[string]int{
			"available": 0, // Days available (will be calculated as inverse of taken)
			"pending":   0, // Days pending approval
			"approved":  0, // Days approved/active
			"rejected":  0, // Days rejected
		}
	}

	// Get user's total leave balance (starting balance)
	var user bson.M
	config.UsersCollection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	totalLeave := 28 // Default starting balance
	if leaveBalance, ok := user["leaveBalance"].(bson.M); ok {
		if total, ok := leaveBalance["total"].(int32); ok {
			totalLeave = int(total)
		}
	}

	// Fill in data from aggregation results
	for _, result := range results {
		idMap := result["_id"].(bson.M)
		monthNum := int(idMap["month"].(int32))
		status := idMap["status"].(string)
		totalDays := 0
		if days, ok := result["totalDays"].(int32); ok {
			totalDays = int(days)
		}

		if monthNum >= 1 && monthNum <= 12 {
			monthKey := months[monthNum-1]
			// Categorize by actual status
			if status == "Pending" {
				monthlyData[monthKey]["pending"] += totalDays
			} else if status == "Approved" || status == "Active" || status == "Over" {
				monthlyData[monthKey]["approved"] += totalDays
			} else if status == "Rejected" {
				monthlyData[monthKey]["rejected"] += totalDays
			}
		}
	}

	// Get current month to determine which months to show
	currentMonth := int(time.Now().Month())

	// Convert to array format with cumulative available calculation
	graphData := []gin.H{}
	runningAvailable := totalLeave // Start with total leave

	for i, month := range months {
		monthNum := i + 1 // 1-based month number

		// Only show data for months up to current month
		if monthNum > currentMonth {
			// Future months - no data yet
			graphData = append(graphData, gin.H{
				"month":     month,
				"available": nil,
				"pending":   nil,
				"approved":  nil,
				"rejected":  nil,
			})
		} else {
			pending := monthlyData[month]["pending"]
			approved := monthlyData[month]["approved"]
			rejected := monthlyData[month]["rejected"]

			// Subtract both pending and approved leaves from running available
			// (since we deduct immediately on creation)
			runningAvailable -= (pending + approved)

			graphData = append(graphData, gin.H{
				"month":     month,
				"available": runningAvailable,
				"pending":   pending,
				"approved":  approved,
				"rejected":  rejected,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    graphData,
	})
}

// GetAllDashboardData returns combined dashboard data in one response
func GetAllDashboardData(c *gin.Context) {
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

	// Get stats
	totalLeaves, _ := config.LeavesCollection.CountDocuments(ctx, bson.M{"employee": userID})
	pendingCount, _ := config.LeavesCollection.CountDocuments(ctx, bson.M{
		"employee": userID,
		"status":   "Pending",
	})
	approvedCount, _ := config.LeavesCollection.CountDocuments(ctx, bson.M{
		"employee": userID,
		"status":   "Approved",
	})
	rejectedCount, _ := config.LeavesCollection.CountDocuments(ctx, bson.M{
		"employee": userID,
		"status":   "Rejected",
	})

	// Get recent leaves
	pipeline := []bson.M{
		{"$match": bson.M{"employee": userID}},
		{"$sort": bson.M{"createdAt": -1}},
		{"$limit": 5},
		{
			"$lookup": bson.M{
				"from":         "users",
				"localField":   "reliever",
				"foreignField": "_id",
				"as":           "relieverData",
			},
		},
		{
			"$unwind": bson.M{
				"path":                       "$relieverData",
				"preserveNullAndEmptyArrays": true,
			},
		},
	}

	cursor, _ := config.LeavesCollection.Aggregate(ctx, pipeline)
	var recentLeaves []bson.M
	cursor.All(ctx, &recentLeaves)
	cursor.Close(ctx)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"stats": gin.H{
				"totalLeaves":    totalLeaves,
				"pendingLeaves":  pendingCount,
				"approvedLeaves": approvedCount,
				"rejectedLeaves": rejectedCount,
			},
			"recentLeaves": recentLeaves,
		},
	})
}
