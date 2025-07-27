package database

import (
	"log"
	"os"
	"time"

	"biblioteca-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB es la instancia global de la base de datos
var DB *gorm.DB

func ConnectDB() {
	var err error

	// Obtener URL de conexión
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=biblioteca password=biblioteca123 dbname=biblioteca_db port=5432 sslmode=disable"
	}

	log.Printf("Conectando a la base de datos con DSN: %s", dsn)

	// Configurar logger
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)

	// Conectar a la base de datos
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})

	if err != nil {
		log.Fatal("Error al conectar con la base de datos:", err)
	}

	log.Println("Conexión a la base de datos exitosa")

	// Configurar pool de conexiones
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Error al obtener SQL DB:", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto migrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.Book{},
		&models.Exemplar{},
		&models.Loan{},
		&models.Fine{},
		&models.Recommendation{},
		&models.LoanHistory{},
		&models.FineHistory{},
	)

	if err != nil {
		log.Fatal("Error en auto migrate:", err)
	}

	log.Println("Auto migrate completado exitosamente")

	// Seed data (opcional)
	seedData()
}

func seedData() {
	// Crear usuario admin por defecto si no existe
	var adminUser models.User
	result := DB.Where("login = ?", "admin").First(&adminUser)

	if result.Error != nil {
		admin := models.User{
			Login:    "admin",
			Name:     "Administrador",
			LastName: "Sistema",
			Email:    "admin@biblioteca.com",
			UserType: models.ADMIN,
			Status:   models.ACTIVE,
		}
		DB.Create(&admin)
		log.Println("Usuario admin creado")
	}
}
