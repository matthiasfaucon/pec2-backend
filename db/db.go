package db

import (
	"fmt"
	"pec2-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	// Configuration de la chaîne de connexion (DSN)
	dsn := "postgresql://neondb_owner:npg_PBaO3fRmU0MY@ep-misty-frost-a51ypwcn-pooler.us-east-2.aws.neon.tech/neondb?sslmode=require"
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		panic("Could not connect to the database")
	}

	// Auto-migrer la table User
	err = DB.AutoMigrate(&models.User{})
	if err != nil {
		fmt.Println("Error migrating database:", err)
		panic("Could not migrate database")
	}

	fmt.Println("Database connection successful")
}
