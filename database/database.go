package database

import (
	"fmt"
	"gin-quickstart/models"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_SSLMODE"),
		os.Getenv("DB_TIMEZONE"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database! Error: ", err)
	}

	log.Println("Successfully connected to the database!")

	log.Println("Running AutoMigrate...")
	err = db.AutoMigrate(
		&models.Patient{},
		&models.Doctor{},
		&models.Examination{},
		&models.Appointment{},
		&models.RefreshToken{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database! Error: ", err)
	}

	DB = db
}
