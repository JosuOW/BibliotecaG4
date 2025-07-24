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

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes
		books := api.Group("/books")
		{
			books.GET("", controllers.GetBooks)
			books.GET("/:id", controllers.GetBookByID)
		}

		// Protected routes (TODO: add auth middleware)
		// protected := api.Group("/")
		// protected.Use(middleware.AuthMiddleware())
		// {
		//     protected.POST("/books", controllers.CreateBook)
		// }
	}

	return r
}
