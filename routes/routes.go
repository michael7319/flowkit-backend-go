package routes

import (
	"github.com/flowkit/backend/handlers"
	"github.com/flowkit/backend/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine) {
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

			// Approver routes (HOD, HR, GED)
			leaves.GET("", middleware.AuthorizeRoles("HOD", "HR", "GED", "admin"), handlers.GetAllLeaves)
			leaves.PUT("/:id/approve", middleware.AuthorizeRoles("HOD", "HR", "GED", "admin"), handlers.ApproveLeave)
			leaves.PUT("/:id/reject", middleware.AuthorizeRoles("HOD", "HR", "GED", "admin"), handlers.RejectLeave)
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
		admin.GET("/users", handlers.GetUsers)
		// Add more admin-specific routes here
	}

	// Health check
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "FlowKit Leave Management API is running",
		})
	})
}
