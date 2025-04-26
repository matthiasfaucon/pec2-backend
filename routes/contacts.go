package routes

import (
	"pec2-backend/handlers/contacts"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func ContactsRoutes(r *gin.Engine) {
	// Route publique - accessible sans authentification
	r.POST("/contact", contacts.CreateContact)

	// Routes protégées
	contactRoutes := r.Group("/contacts")
	contactRoutes.Use(middleware.JWTAuth())
	{
		// Routes accessibles uniquement aux administrateurs
		contactRoutes.GET("", middleware.AdminAuth(), contacts.GetAllContacts)
		contactRoutes.PUT("/:id/status", middleware.AdminAuth(), contacts.UpdateContactStatus)
	}
}
