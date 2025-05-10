package routes

import (
	"pec2-backend/handlers/posts"
	"pec2-backend/handlers/posts/likes"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func PostsRoutes(r *gin.Engine) {
	// Routes publiques
	r.GET("/posts", posts.GetAllPosts)
	r.GET("/posts/:id", posts.GetPostByID)

	// Routes protégées
	postsRoutes := r.Group("/posts")
	postsRoutes.Use(middleware.JWTAuth())
	{
		postsRoutes.POST("", posts.CreatePost)
		postsRoutes.PUT("/:id", posts.UpdatePost)
		postsRoutes.DELETE("/:id", posts.DeletePost)

		// Routes des interactions
		postsRoutes.POST("/:id/like", likes.ToggleLike)
	}
}
