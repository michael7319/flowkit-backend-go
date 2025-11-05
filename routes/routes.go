package routes

import (
	"github.com/flowkit/backend/handlers"
	"github.com/flowkit/backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
	// Root endpoint - for Render health checks
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "FlowKit API is running",
			"version": "1.0.0",
		})
	})

	// API prefix
	api := router.Group("/api")

	// Public routes - Authentication
	auth := api.Group("/auth")
	{
		auth.POST("/register", handlers.Register)
		auth.POST("/login", handlers.Login)
	}

	// Protected routes - require authentication
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// Auth routes (authenticated)
		authProtected := protected.Group("/auth")
		{
			authProtected.GET("/me", handlers.GetMe)
			authProtected.PUT("/update-password", handlers.UpdatePassword)
		}

		// User routes
		users := protected.Group("/users")
		{
			users.GET("", handlers.GetUsers)
			users.GET("/relievers", handlers.GetRelievers)
			users.GET("/:id", handlers.GetUserByID)
			users.PUT("/profile", handlers.UpdateProfile)
			users.POST("/signature", handlers.UploadSignature)
		}

		// Leave routes
		leaves := protected.Group("/leaves")
		{
			// Employee routes
			leaves.POST("", handlers.CreateLeave)
			leaves.GET("/my-leaves", handlers.GetMyLeaves)
			leaves.PUT("/:id", handlers.UpdateLeave)
			leaves.DELETE("/:id", handlers.DeleteLeave)
			leaves.PUT("/:id/cancel", handlers.CancelLeave)

			// General approver routes (for backward compatibility)
			leaves.GET("", middleware.AuthorizeRoles("hod", "hr", "ged", "admin"), handlers.GetAllLeaves)
			leaves.PUT("/:id/approve", middleware.AuthorizeRoles("hod", "hr", "ged", "admin"), handlers.ApproveLeave)
			leaves.PUT("/:id/reject", middleware.AuthorizeRoles("hod", "hr", "ged", "admin"), handlers.RejectLeave)
		}

		// HOD-specific approval routes
		hod := protected.Group("/hod")
		hod.Use(middleware.AuthorizeRoles("hod", "admin"))
		{
			hod.GET("/leaves", handlers.GetHODLeaves)                // Get department leaves
			hod.PUT("/leaves/:id/approve", handlers.HODApproveLeave) // HOD approve
			hod.PUT("/leaves/:id/reject", handlers.HODRejectLeave)   // HOD reject
		}

		// HR-specific approval routes
		hr := protected.Group("/hr")
		hr.Use(middleware.AuthorizeRoles("hr", "admin"))
		{
			hr.GET("/leaves", handlers.GetHRLeaves)                // Get HOD-approved leaves
			hr.PUT("/leaves/:id/approve", handlers.HRApproveLeave) // HR approve
			hr.PUT("/leaves/:id/reject", handlers.HRRejectLeave)   // HR reject
		}

		// GED-specific approval routes
		ged := protected.Group("/ged")
		ged.Use(middleware.AuthorizeRoles("ged", "admin"))
		{
			ged.GET("/leaves", handlers.GetGEDLeaves)                // Get HR-approved leaves
			ged.PUT("/leaves/:id/approve", handlers.GEDApproveLeave) // GED approve (final)
			ged.PUT("/leaves/:id/reject", handlers.GEDRejectLeave)   // GED reject
		}

		// Dashboard routes
		dashboard := protected.Group("/dashboard")
		{
			dashboard.GET("/stats", handlers.GetDashboardStats)
			dashboard.GET("/progress/:id", handlers.GetLeaveProgress)
			dashboard.GET("/graph", handlers.GetGraphData)
			dashboard.GET("/all", handlers.GetAllDashboardData)
		}
	}

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(middleware.AuthMiddleware(), middleware.AuthorizeRoles("admin"))
	{
		// User Management
		admin.POST("/users", handlers.AdminCreateUser)                              // Create user
		admin.GET("/users", handlers.AdminGetAllUsers)                              // Get all users (with filters)
		admin.PUT("/users/:id", handlers.AdminUpdateUser)                           // Update user info
		admin.PUT("/users/:id/activate", handlers.AdminActivateUser)                // Activate user
		admin.PUT("/users/:id/deactivate", handlers.AdminDeactivateUser)            // Deactivate user
		admin.PUT("/users/:id/password", handlers.AdminResetUserPassword)           // Reset password
		admin.PUT("/users/:id/leave-balance", handlers.AdminUpdateUserLeaveBalance) // Update leave balance
	}

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "FlowKit Leave Management API is running",
		})
	})
}
