package routes

import (
	"pec2-backend/handlers/content_creators"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func ContentCreatorsRoutes(r *gin.Engine) {
	contentCreatorRoutes := r.Group("/content-creators")
	contentCreatorRoutes.Use(middleware.JWTAuth())
	{
		// Routes publiques (accessibles à tous les utilisateurs authentifiés)
		contentCreatorRoutes.POST("", content_creators.Apply)
		contentCreatorRoutes.PUT("", content_creators.UpdateContentCreatorInfo)

		// Routes admin
		contentCreatorRoutes.GET("/all", middleware.AdminAuth(), content_creators.GetAllContentCreators)
	}
}
