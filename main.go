package main

import (
	"log"

	"pec2-backend/db"
	_ "pec2-backend/docs"
	"pec2-backend/routes"
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
)

// @title API PEC2 Backend
// @version 1.0
// @description API pour le projet PEC2 Backend
// @host localhost:8080
// @BasePath /
// @SecurityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Entrez le JWT avec le préfixe Bearer: Bearer <JWT>
func main() {
	
	gin.SetMode(gin.ReleaseMode)

	// Initialiser la base de données
	db.InitDB()

	// Initialiser Cloudinary
	if err := utils.InitCloudinary(); err != nil {
		log.Printf("Error while initializing Cloudinary: %v", err)
		log.Println("The image upload will not work correctly.")
	}

	r := routes.SetupRouter()


	if err := r.Run("0.0.0.0:8090"); err != nil {
		log.Fatal("Error while starting the server:", err)
	}
}
