package routes

import (
	"pec2-backend/handlers/users"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func UsersRoutes(r *gin.Engine) {
	r.GET("/users", users.GetAllUsers)
	r.GET("/users/:id", users.GetUserByID)

	protected := r.Group("")
	protected.Use(middleware.JWTAuth())
	{
		protected.PUT("/users/password", users.UpdatePassword)
	}
}
