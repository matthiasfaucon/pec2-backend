package main

import (
	"log"
	"os"

	"pec2-backend/db"
	"pec2-backend/docs"
	"pec2-backend/routes"
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
)

// @title API PEC2 Backend
// @version 1.0
// @description API pour le projet PEC2 Backend
// @host localhost:8090
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

	// Récupérer les variables d'environnement
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	docs.SwaggerInfo.Host = "localhost:" + port

	log.Printf("Server is running on port %s\n", port)
	r := routes.SetupRouter()

	if err := r.Run(baseURL + ":" + port); err != nil {
		log.Fatal("Error while starting the server:", err)
	}
}
