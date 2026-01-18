package routes

import (
	"DompetKu/controllers"
	"DompetKu/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine) {
	// Public routes
	api := router.Group("/api")
	{
		// Auth routes (no auth required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// Get categories (public)
		api.GET("/categories", controllers.GetCategories)

		// Protected routes
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware())
		{
			// User profile routes
			user := protected.Group("/user")
			{
				user.GET("/profile", controllers.GetProfile)
				user.PUT("/profile", controllers.UpdateProfile)
				user.PUT("/change-password", controllers.ChangePassword)
			}

			// Transaction routes
			transactions := protected.Group("/transactions")
			{
				transactions.POST("", controllers.CreateTransaction)
				transactions.GET("", controllers.GetTransactions)
				transactions.GET("/:id", controllers.GetTransactionByID)
				transactions.PUT("/:id", controllers.UpdateTransaction)
				transactions.DELETE("/:id", controllers.DeleteTransaction)
			}

			// Financial Goals routes
			goals := protected.Group("/goals")
			{
				goals.POST("", controllers.CreateGoal)
				goals.GET("", controllers.GetGoals)
				goals.GET("/:id", controllers.GetGoalByID)
				goals.PUT("/:id", controllers.UpdateGoal)
				goals.POST("/:id/add", controllers.AddProgressToGoal)
				goals.DELETE("/:id", controllers.DeleteGoal)
			}

			// Statistics routes
			stats := protected.Group("/stats")
			{
				stats.GET("/summary", controllers.GetSummary)
				stats.GET("/expense-by-category", controllers.GetExpenseByCategory)
				stats.GET("/income-vs-expense", controllers.GetIncomeVsExpense)
			}
		}
	}
}
