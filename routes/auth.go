package routes

import (
	"pec2-backend/handlers/auth"

	"github.com/gin-gonic/gin"
)

func AuthRoutes(r *gin.Engine) {
	r.POST("/register", auth.CreateUser)
	r.POST("/login", auth.Login)
	r.GET("/valid-email/:code", auth.ValidEmail)
}
