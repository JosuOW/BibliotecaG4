package routes

import (
	"biblioteca-backend/controllers"
	"biblioteca-backend/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRoutes() *gin.Engine {
	r := gin.New()

	// Middleware global
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", controllers.HealthCheck)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Public routes
	api := r.Group("/api/v1")
	{
		api.GET("/books", controllers.GetBooks)
		api.GET("/books/:id", controllers.GetBookByID)
	}

	// Protected routes
	protected := api.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		// User routes
		protected.GET("/loans/my", controllers.GetMyLoans)
		protected.GET("/loans/history", controllers.GetLoanHistory) // ✅ ← Agregada
		protected.POST("/loans", controllers.CreateLoan)
		protected.PUT("/loans/:id/return", controllers.ReturnLoan)

		// Admin routes
		admin := protected.Group("/admin")
		admin.Use(middleware.RequireRole("ADMIN"))
		{
			admin.POST("/users", controllers.CreateUser)
			admin.POST("/books", controllers.CreateBook)
			admin.GET("/users", controllers.GetAllUsers)
		}
	}

	return r
}
