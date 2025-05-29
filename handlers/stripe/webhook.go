package stripe

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"pec2-backend/db"
	"pec2-backend/models"

	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func StripeWebhookHandler(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	payload, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	// Vérification de la signature Stripe
	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	fmt.Println("Webhook secret utilisé :", endpointSecret)
	sigHeader := c.GetHeader("Stripe-Signature")
	fmt.Println("Signature header reçu :", sigHeader)
	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		fmt.Println("Erreur signature Stripe:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature verification failed"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing webhook JSON"})
			return
		}
		// 1. Récupérer le user correspondant au StripeCustomerId
		var user models.User
		err := db.DB.First(&user, "stripe_customer_id = ?", session.Customer.ID).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found for this Stripe customer"})
			return
		}
		// 2. Vérifier que le content creator existe
		var creator models.User
		err = db.DB.First(&creator, "id = ?", session.ClientReferenceID).Error
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Content creator not found"})
			return
		}
		if creator.Role != models.ContentCreator {
			c.JSON(http.StatusForbidden, gin.H{"error": "Target is not a content creator"})
			return
		}
		// 3. Créer l'abonnement en base
		stripeSubID := ""
		if session.Subscription != nil {
			stripeSubID = session.Subscription.ID
		}
		sub := models.Subscription{
			UserID:               user.ID,
			ContentCreatorID:     creator.ID,
			Status:               models.SubscriptionActive,
			StripeSubscriptionId: stripeSubID,
			StartDate:            time.Now(),
		}
		err = db.DB.Create(&sub).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création de l'abonnement"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Abonnement créé en base"})
		return
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event ignoré"})
	}
}
