package routes

import (
	"pec2-backend/handlers/privateMessages"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func PrivateMessagesRoutes(r *gin.Engine) {
	privateMessagesGroup := r.Group("/private-messages")
	privateMessagesGroup.Use(middleware.JWTAuth())
	{
		privateMessagesGroup.POST("", privateMessages.CreatePrivateMessage)
		privateMessagesGroup.GET("", privateMessages.GetUserMessages)
	}
}
