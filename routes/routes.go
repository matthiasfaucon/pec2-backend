package routes

import (
	"pec2-backend/handlers/ping"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configure toutes les routes de l'application
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Initialisation des handlers
	pingHandler := ping.New()

	// Documentation Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Routes
	r.GET("/ping", pingHandler.HandlePing)

	return r
}


