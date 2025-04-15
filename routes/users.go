package routes

import (
	"pec2-backend/handlers/users"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func UsersRoutes(r *gin.Engine) {
	
	// Route accessible sans authentification
	r.GET("/users/:id", users.GetUserByID)

	userRoutes := r.Group("/users")
	userRoutes.Use(middleware.JWTAuth())
	{
		// Route accessible uniquement aux administrateurs
		userRoutes.GET("", middleware.AdminAuth(), users.GetAllUsers)

		// Routes accessibles à tout utilisateur authentifié
		userRoutes.PUT("/password", users.UpdatePassword)
	}
}
