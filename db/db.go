package db

import (
	"fmt"
	"os"
	"pec2-backend/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Tentative de chargement du fichier .env
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: Impossible to load the .env file:", err)
		fmt.Println("The environment variable DB_URL must be defined in the system environment")
	}

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		fmt.Println("Variable DB_URL non définie")
		panic("URL de base de données non configurée")
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		panic("Could not connect to the database")
	}

	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		fmt.Println("Error migrating database:", err)
		panic("Could not migrate database")
	}

	fmt.Println("Database connection successful")
}
