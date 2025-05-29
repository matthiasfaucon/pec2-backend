package stripe

import (
	"encoding/json"
	"fmt"
	"io"
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
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("Error reading request body: %v\n", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Error reading request body"})
		return
	}

	endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if endpointSecret == "" {
		fmt.Println("ERREUR: STRIPE_WEBHOOK_SECRET non défini")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook secret not configured"})
		return
	}

	fmt.Println("Webhook secret used:", endpointSecret)
	sigHeader := c.GetHeader("Stripe-Signature")
	fmt.Println("Received signature header:", sigHeader)

	event, err := webhook.ConstructEvent(payload, sigHeader, endpointSecret)
	if err != nil {
		fmt.Printf("Erreur signature Stripe: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Signature verification failed"})
		return
	}

	fmt.Printf("Webhook received - Type: %s, ID: %s\n", event.Type, event.ID)

	switch event.Type {
	case "checkout.session.completed":
		handleCheckoutSessionCompleted(c, event)
	case "payment_intent.succeeded":
		handlePaymentIntentSucceeded(c, event)
	case "payment_intent.payment_failed":
		handlePaymentIntentFailed(c, event)
	default:
		fmt.Printf("Unhandled event type: %s\n", event.Type)
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
	}
}

func handleCheckoutSessionCompleted(c *gin.Context, event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		fmt.Printf("Erreur parsing checkout session: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing checkout session JSON"})
		return
	}

	fmt.Printf("Checkout session completed - ID: %s, Customer: %s\n", session.ID, session.Customer.ID)

	if session.Customer == nil || session.Customer.ID == "" {
		fmt.Println("Customer ID missing in session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer ID missing in session"})
		return
	}

	if session.ClientReferenceID == "" {
		fmt.Println("ClientReferenceID missing in session")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Content creator ID missing"})
		return
	}

	var user models.User
	err := db.DB.First(&user, "stripe_customer_id = ?", session.Customer.ID).Error
	if err != nil {
		fmt.Printf("User not found for customer ID: %s, error: %v\n", session.Customer.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found for this Stripe customer"})
		return
	}

	var creator models.User
	err = db.DB.First(&creator, "id = ?", session.ClientReferenceID).Error
	if err != nil {
		fmt.Printf("Content creator not found for ID: %s, error: %v\n", session.ClientReferenceID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Content creator not found"})
		return
	}

	if creator.Role != models.ContentCreator {
		fmt.Printf("The user %s is not a content creator (role: %v)\n", creator.ID, creator.Role)
		c.JSON(http.StatusForbidden, gin.H{"error": "Target is not a content creator"})
		return
	}

	stripeSubID := ""
	if session.Subscription != nil {
		stripeSubID = session.Subscription.ID
		var existingSub models.Subscription
		err = db.DB.First(&existingSub, "stripe_subscription_id = ?", stripeSubID).Error
		if err == nil {
			fmt.Printf("Subscription already exists for ID: %s\n", stripeSubID)
			c.JSON(http.StatusOK, gin.H{"message": "Subscription already exists"})
			return
		}
	}

	var existingUserSub models.Subscription
	err = db.DB.Where("user_id = ? AND content_creator_id = ? AND status = ?",
		user.ID, creator.ID, models.SubscriptionActive).First(&existingUserSub).Error
	if err == nil {
		fmt.Printf("Active subscription already exists between user %s and creator %s\n", user.ID, creator.ID)
		c.JSON(http.StatusConflict, gin.H{"error": "Active subscription already exists between these users"})
		return
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, 1, 0)

	sub := models.Subscription{
		UserID:               user.ID,
		ContentCreatorID:     creator.ID,
		Status:               models.SubscriptionPending,
		StripeSubscriptionId: stripeSubID,
		StartDate:            startDate,
		EndDate:              &endDate,
	}

	err = db.DB.Create(&sub).Error
	if err != nil {
		fmt.Printf("Erreur création subscription: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating subscription"})
		return
	}

	fmt.Printf("Subscription created successfully: ID=%s, UserID=%s, CreatorID=%s, Status=Pending\n",
		sub.ID, user.ID, creator.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Subscription created successfully, waiting for payment confirmation"})
}

func handlePaymentIntentSucceeded(c *gin.Context, event stripe.Event) {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		fmt.Printf("Erreur parsing payment intent: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing payment intent JSON"})
		return
	}

	fmt.Printf("Payment intent succeeded - ID: %s, Amount: %d, Customer: %s\n",
		paymentIntent.ID, paymentIntent.Amount, paymentIntent.Customer.ID)

	var existingPayment models.SubscriptionPayment
	err := db.DB.First(&existingPayment, "stripe_payment_intent_id = ?", paymentIntent.ID).Error
	if err == nil {
		fmt.Printf("Payment already recorded for PaymentIntent: %s\n", paymentIntent.ID)
		c.JSON(http.StatusOK, gin.H{"message": "Payment already recorded"})
		return
	}

	var user models.User
	err = db.DB.First(&user, "stripe_customer_id = ?", paymentIntent.Customer.ID).Error
	if err != nil {
		fmt.Printf("User not found for customer ID: %s\n", paymentIntent.Customer.ID)
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found for this payment"})
		return
	}

	var subscription models.Subscription
	err = db.DB.Where("user_id = ? AND status IN (?)",
		user.ID, []models.SubscriptionStatus{models.SubscriptionPending, models.SubscriptionActive}).
		Order("created_at DESC").First(&subscription).Error
	if err != nil {
		fmt.Printf("Subscription non trouvée pour user ID: %s\n", user.ID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found for this payment"})
		return
	}

	amount := int(paymentIntent.Amount)
	payment := models.SubscriptionPayment{
		SubscriptionID:        subscription.ID,
		Amount:                amount,
		PaidAt:                time.Now(),
		StripePaymentIntentId: paymentIntent.ID,
	}

	err = db.DB.Create(&payment).Error
	if err != nil {
		fmt.Printf("Erreur création du paiement: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating payment record"})
		return
	}

	err = db.DB.Model(&subscription).Update("status", models.SubscriptionActive).Error
	if err != nil {
		fmt.Printf("Erreur activation de la subscription: %v\n", err)
	} else {
		fmt.Printf("Subscription %s activée avec succès\n", subscription.ID)
	}

	fmt.Printf("Paiement créé avec succès: Amount=%d€, SubscriptionID=%s, PaymentIntentID=%s\n",
		amount, subscription.ID, paymentIntent.ID)
	c.JSON(http.StatusOK, gin.H{"message": "Payment recorded and subscription activated"})
}

func handlePaymentIntentFailed(c *gin.Context, event stripe.Event) {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		fmt.Printf("Erreur parsing payment intent failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing payment intent failed JSON"})
		return
	}

	errorMessage := "Unknown error"
	if paymentIntent.LastPaymentError != nil {
		errorMessage = paymentIntent.LastPaymentError.Msg
	}

	fmt.Printf("Échec de paiement - PaymentIntent ID: %s, Customer: %s, Raison: %s\n",
		paymentIntent.ID, paymentIntent.Customer.ID, errorMessage)

	if paymentIntent.Customer != nil {
		var user models.User
		err := db.DB.First(&user, "stripe_customer_id = ?", paymentIntent.Customer.ID).Error
		if err == nil {
			err = db.DB.Model(&models.Subscription{}).
				Where("user_id = ? AND status = ? AND created_at > ?",
					user.ID, models.SubscriptionPending, time.Now().Add(-1*time.Hour)).
				Update("status", models.SubscriptionCanceled).Error
			if err != nil {
				fmt.Printf("Error updating pending subscriptions: %v\n", err)
			} else {
				fmt.Printf("Pending subscriptions marked as canceled for user %s\n", user.ID)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Payment failure processed"})
}
