package routes

import (
	"pec2-backend/handlers/stripe"
	"pec2-backend/middleware"

	"github.com/gin-gonic/gin"
)

func StripeRoutes(r *gin.Engine) {
	subscriptionRoutes := r.Group("/subscriptions")
	subscriptionRoutes.Use(middleware.JWTAuth())
	{
		subscriptionRoutes.POST("/checkout/:contentCreatorId", stripe.CreateSubscriptionCheckoutSession)
		subscriptionRoutes.DELETE("/:subscriptionId", stripe.CancelSubscription)
		subscriptionRoutes.GET("/", stripe.GetUserSubscriptions)
		subscriptionRoutes.GET("/:subscriptionId", stripe.GetSubscriptionDetail)
	}
	r.POST("/stripe/webhook", stripe.StripeWebhookHandler)
}
