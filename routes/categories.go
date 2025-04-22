package routes

import (
	"pec2-backend/handlers/categories"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func CategoriesRoutes(r *gin.Engine) {	
	// Routes protégées mais publiques
	categoriesPublicRoutes := r.Group("/categories")
	categoriesPublicRoutes.Use(middleware.JWTAuth())
	categoriesPublicRoutes.GET("", categories.GetAllCategories)

	// Routes des catégories protégées (admin seulement)
	categoriesPrivateRoutes := r.Group("/categories")
	categoriesPrivateRoutes.Use(middleware.JWTAuth())
	categoriesPrivateRoutes.Use(middleware.AdminAuth())
	{
		categoriesPrivateRoutes.POST("", categories.CreateCategory)
		categoriesPrivateRoutes.PUT("/:id", categories.UpdateCategory)
		categoriesPrivateRoutes.DELETE("/:id", categories.DeleteCategory)
	}
}
