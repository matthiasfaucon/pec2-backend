package routes

import (
	"pec2-backend/handlers/users"

	"github.com/gin-gonic/gin"
)

// UsersRoutes définit les routes liées aux utilisateurs
func UsersRoutes(r *gin.Engine) {
	r.GET("/users", users.GetAllUsers)
	r.GET("/users/:id", users.GetUserByID)
}
