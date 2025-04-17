package main

import (
	"log"

	"pec2-backend/db"
	_ "pec2-backend/docs"
	"pec2-backend/routes"
	"pec2-backend/utils"

	"time"

	"github.com/gin-contrib/cors"
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
		log.Printf("Avertissement: Initialisation de Cloudinary a échoué: %v", err)
		log.Println("Le téléchargement d'images ne fonctionnera pas correctement.")
	}

	r := routes.SetupRouter()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Pour autoriser toutes les origines en dev
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
