package database

import (
	"fmt"
	"log"
	"os"

	"QR-GENERATOR/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDB initializes the database connection and performs migrations
func InitDB() error {
	host := os.Getenv("DATABASE_HOST")
	port := os.Getenv("DATABASE_PORT")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	sslmode := os.Getenv("DATABASE_SSLMODE")

	if sslmode == "" {
		sslmode = "disable"
	}

	// Build DSN (Data Source Name)
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return err
	}

	log.Println("✓ Database connected successfully")

	// Auto-migrate all models
	err = DB.AutoMigrate(
		&models.Location{},
		&models.Item{},
		&models.User{},
		&models.ItemMovement{},
	)

	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
		return err
	}

	log.Println("✓ Database migrations completed")
	return nil
}

// GetDB returns the database connection
func GetDB() *gorm.DB {
	return DB
}
