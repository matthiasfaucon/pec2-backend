package main

import (
	"log"

	_ "pec2-backend/docs"
	"pec2-backend/routes"

	"github.com/gin-gonic/gin"
)

// @title API PEC2 Backend
// @version 1.0
// @description API pour le projet PEC2 Backend
// @host localhost:8080
// @BasePath /
func main() {
	// Initialiser Gin en mode release
	gin.SetMode(gin.ReleaseMode)

	// Initialiser le router avec les routes
	r := routes.SetupRouter()

	// Démarrer le serveur
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
