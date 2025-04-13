package routes

import (
	"pec2-backend/handlers/contacts"

	"github.com/gin-gonic/gin"
)

func ContactsRoutes(r *gin.Engine) {
	r.POST("/contact", contacts.CreateContact)
}
