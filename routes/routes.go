package routes

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// SetupRouter configure toutes les routes de l'application
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Documentation Swagger
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Route de test ping
	r.GET("/ping", Ping)

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
