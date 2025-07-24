package main

import (
	"log"
	"os"

	"biblioteca-backend/database"
	"biblioteca-backend/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// @title Biblioteca API
// @version 1.0
// @description API para sistema de gestión de biblioteca
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8000
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Cargar variables de entorno
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found")
	}

	// Conectar a la base de datos
	database.ConnectDB()

	// Configurar modo Gin
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Configurar rutas
	r := routes.SetupRoutes()

	// Obtener puerto
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Servidor iniciado en puerto %s", port)
	log.Fatal(r.Run(":" + port))
}
