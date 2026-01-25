package api

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"DompetKu/config"
	"DompetKu/controllers"
	"DompetKu/middleware"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ginEngine *gin.Engine
	once      sync.Once
)

func initDB() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Println("Warning: MONGO_URI not set")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Println("Failed to connect to MongoDB:", err)
		return
	}

	dbName := os.Getenv("MONGO_DB")
	if dbName == "" {
		dbName = "dompetku_db"
	}
	
	// Set the DB in config package so controllers can use it
	config.SetDB(client.Database(dbName))
	log.Println("Connected to MongoDB!")
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Public routes
	api := router.Group("/api")
	{
		// Auth routes
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
				goals.POST("/:id/withdraw", controllers.WithdrawFromGoal)
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

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "DompetKu API is running on Vercel",
		})
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to DompetKu API",
			"version": "1.0.0",
		})
	})

	return router
}

func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		initDB()
		ginEngine = setupRouter()
	})

	ginEngine.ServeHTTP(w, r)
}
