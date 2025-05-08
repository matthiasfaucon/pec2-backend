package routes

import (
	"pec2-backend/handlers/content_creators"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func ContentCreatorsRoutes(r *gin.Engine) {
	// Routes protégées nécessitant une authentification
	contentCreatorRoutes := r.Group("/content-creators")
	contentCreatorRoutes.Use(middleware.JWTAuth())
	{
		contentCreatorRoutes.POST("/", content_creators.Apply)
	}


}
