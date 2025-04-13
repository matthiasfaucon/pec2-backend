package routes

import (
	"pec2-backend/handlers/users"

	"github.com/gin-gonic/gin"
)

func UsersRoutes(r *gin.Engine) {
	r.POST("/user", users.CreateUser)
}
