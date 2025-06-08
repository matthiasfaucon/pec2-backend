package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Pour autoriser toutes les origines en dev
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/", test)
	AuthRoutes(r)
	ContactsRoutes(r)
	UsersRoutes(r)
	CategoriesRoutes(r)
	PostsRoutes(r)
	ContentCreatorsRoutes(r)
	InseeRoutes(r)
	PrivateMessagesRoutes(r)
	StripeRoutes(r)

	return r
}
func test(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Bienvenue CD factorisé !",
	})
}
