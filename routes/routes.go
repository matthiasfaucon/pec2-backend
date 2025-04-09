package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"pec2-backend/handlers/ping"
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

// @Summary Ping test
// @Description Endpoint de test qui répond pong
// @Tags test
// @Produce json
// @Success 200 {object} map[string]string
// @Router /ping [get]
func Ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
