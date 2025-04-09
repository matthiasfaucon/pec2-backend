package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialiser Gin en mode release
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Routes
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// Démarrer le serveur
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Erreur lors du démarrage du serveur:", err)
	}
}
