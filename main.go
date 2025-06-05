package main

import (
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

	// Fais en sorte que les logs de Gin et les logs logrus soient dans le même format
	gin.DefaultWriter = utils.LogWriter()
	gin.DefaultErrorWriter = utils.LogWriter()

	// Possibilité de supprimer les logs de Gin
	gin.DisableConsoleColor()

	// Initialiser Cloudinary
	if err := utils.InitCloudinary(); err != nil {
		utils.LogError(err, "Error when initializing Cloudinary")
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

	utils.LogSuccess("Le serveur fonctionne sur le port " + port + " baseUrl:" + baseURL)
	r := routes.SetupRouter()

	if err := r.Run(baseURL + ":" + port); err != nil {
		utils.LogError(err, "Error when starting the server")
	}
}
