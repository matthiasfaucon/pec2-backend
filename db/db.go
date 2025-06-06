package db

import (
	"os"
	"pec2-backend/models"
	"pec2-backend/utils"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	if err := godotenv.Load(); err != nil {
		utils.LogError(err, "Warning: Impossible to load the .env file")
		utils.LogInfo("The environment variable DB_URL must be defined in the system environment")
	}

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		utils.LogError(nil, "Variable DB_URL non définie")
		panic("URL de base de données non configurée")
	}

	var err error
	// Utilisation du logger GORM harmonisé
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: utils.GetGormLogger(),
	})
	if err != nil {
		utils.LogError(err, "Error connecting to the database")
		panic("Could not connect to the database")
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Contact{},
		&models.Post{},
		&models.Like{},
		&models.Report{},
		&models.Comment{},
		&models.Category{},
		&models.ContentCreatorInfo{},
		&models.PrivateMessage{},
		&models.Subscription{},
		&models.SubscriptionPayment{},
	)
	if err != nil {
		utils.LogError(err, "Error migrating database")
		panic("Could not migrate database")
	}

	utils.LogSuccess("Database connection successful")
}