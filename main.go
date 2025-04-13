package main

import (
	"log"

	"pec2-backend/db"
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
	gin.SetMode(gin.ReleaseMode)

	db.InitDB()

	r := routes.SetupRouter()

	if err := r.Run(":8080"); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
