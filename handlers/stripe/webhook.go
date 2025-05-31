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
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

func StripeWebhookHandler(c *gin.Context) {
	const MaxBodyBytes = int64(65536)
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		utils.LogError(err, "Impossible to read request body dans StripeWebhookHandler")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Impossible to read request body"})
		return
	}

	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		utils.LogError(nil, "Webhook secret not configured dans StripeWebhookHandler")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook secret not configured"})
		return
	}

	sig := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sig, secret)
	if err != nil {
		utils.LogError(err, "Stripe signature verification failed dans StripeWebhookHandler")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Stripe signature verification failed"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		handleCheckoutSessionCompleted(c, event)
	case "payment_intent.created":
		handlePaymentIntentCreated(c, event)
	case "payment_intent.processing":
		handlePaymentIntentProcessing(c, event)
	case "payment_intent.succeeded":
		handlePaymentIntentSucceeded(c, event)
	case "payment_intent.failed":
		handlePaymentIntentFailed(c, event)
	case "payment_intent.canceled":
		handlePaymentIntentCanceled(c, event)
	case "invoice.payment_succeeded":
		handleInvoicePaymentSucceeded(c, event)
	case "invoice.payment_failed":
		handleInvoicePaymentFailed(c, event)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
	}
}

func handleCheckoutSessionCompleted(c *gin.Context, event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		utils.LogError(err, "Error parsing CheckoutSession dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing CheckoutSession"})
		return
	}

	if session.Customer == nil {
		utils.LogError(nil, "Customer missing in session dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer missing in session"})
		return
	}

	customerID := session.Customer.ID
	creatorID := session.ClientReferenceID
	if creatorID == "" {
		utils.LogError(nil, "ClientReferenceID missing dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusBadRequest, gin.H{"error": "ClientReferenceID missing"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "stripe_customer_id = ?", customerID).Error; err != nil {
		utils.LogError(err, "User not found for this customer dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found for this customer"})
		return
	}

	var creator models.User
	if err := db.DB.First(&creator, "id = ?", creatorID).Error; err != nil {
		utils.LogError(err, "Creator not found dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusNotFound, gin.H{"error": "Creator not found"})
		return
	}

	if creator.Role != models.ContentCreator {
		utils.LogError(nil, "The target is not a content creator dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusForbidden, gin.H{"error": "The target is not a content creator"})
		return
	}

	var stripeSubID string
	if session.Subscription != nil {
		stripeSubID = session.Subscription.ID
		var tmp models.Subscription
		if err := db.DB.First(&tmp, "stripe_subscription_id = ?", stripeSubID).Error; err == nil {
			utils.LogError(nil, "Stripe subscription already exists dans handleCheckoutSessionCompleted")
			c.JSON(http.StatusOK, gin.H{"message": "Stripe subscription already exists"})
			return
		}
	}

	var dup models.Subscription
	if err := db.DB.Where("user_id = ? AND content_creator_id = ? AND status IN ?",
		user.ID, creator.ID,
		[]models.SubscriptionStatus{models.SubscriptionPending, models.SubscriptionActive}).
		First(&dup).Error; err == nil {
		utils.LogError(nil, "Local subscription already exists dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusOK, gin.H{"message": "Local subscription already exists"})
		return
	}

	now := time.Now()
	end := now.AddDate(0, 1, 0)

	initialStatus := models.SubscriptionPending
	if session.PaymentStatus == "paid" {
		initialStatus = models.SubscriptionActive
	}

	sub := models.Subscription{
		UserID:               user.ID,
		ContentCreatorID:     creator.ID,
		Status:               initialStatus,
		StripeSubscriptionId: stripeSubID,
		StartDate:            now,
		EndDate:              &end,
	}

	if err := db.DB.Create(&sub).Error; err != nil {
		utils.LogError(err, "Error creating subscription dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating subscription"})
		return
	}

	if session.PaymentIntent != nil {
		utils.LogError(nil, "PaymentIntent présent dans handleCheckoutSessionCompleted")
		err1 := upsertSubscriptionPayment(sub.ID, int(session.AmountTotal), session.PaymentIntent.ID, models.SubscriptionPaymentPending)
		if err1 != nil {
			utils.LogError(err1, "Erreur upsertSubscriptionPayment (pending) dans handleCheckoutSessionCompleted")
		}
		if session.PaymentStatus == "paid" {
			err2 := upsertSubscriptionPayment(sub.ID, int(session.AmountTotal), session.PaymentIntent.ID, models.SubscriptionPaymentSucceeded)
			if err2 != nil {
				utils.LogError(err2, "Erreur upsertSubscriptionPayment (paid) dans handleCheckoutSessionCompleted")
			}
		}
	} else {
		utils.LogError(nil, "Pas de PaymentIntent dans la session dans handleCheckoutSessionCompleted")
	}

	if initialStatus == models.SubscriptionActive {
		utils.LogSuccess("Subscription created and activated dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusOK, gin.H{"message": "Subscription created and activated"})
	} else {
		utils.LogSuccess("Subscription created, waiting for payment dans handleCheckoutSessionCompleted")
		c.JSON(http.StatusOK, gin.H{"message": "Subscription created, waiting for payment"})
	}
}

func findSubscriptionByCustomer(customerID string, allowMultiple bool) (*models.Subscription, error) {
	var user models.User
	if err := db.DB.First(&user, "stripe_customer_id = ?", customerID).Error; err != nil {
		return nil, fmt.Errorf("user not found")
	}

	var sub models.Subscription
	query := db.DB.Where("user_id = ? AND status IN ?",
		user.ID,
		[]models.SubscriptionStatus{models.SubscriptionPending, models.SubscriptionActive})

	if !allowMultiple {
		query = query.Order("created_at desc")
	}

	if err := query.First(&sub).Error; err != nil {
		return nil, fmt.Errorf("subscription not found")
	}

	return &sub, nil
}

func findSubscriptionByStripeID(stripeSubID string) (*models.Subscription, error) {
	var sub models.Subscription
	if err := db.DB.First(&sub, "stripe_subscription_id = ?", stripeSubID).Error; err != nil {
		return nil, fmt.Errorf("subscription not found")
	}
	return &sub, nil
}

func upsertSubscriptionPayment(subscriptionID string, amount int, paymentIntentID string, status models.SubscriptionPaymentStatus) error {
	if paymentIntentID == "" {
		return nil
	}

	var payment models.SubscriptionPayment
	err := db.DB.First(&payment, "stripe_payment_intent_id = ?", paymentIntentID).Error

	if err == nil {
		// Le paiement existe déjà
		if payment.Status == models.SubscriptionPaymentSucceeded && status == models.SubscriptionPaymentSucceeded {
			// Éviter de mettre à jour un paiement déjà réussi
			return fmt.Errorf("payment already recorded")
		}

		// Mettre à jour uniquement si le nouveau statut est différent
		if payment.Status != status {
			return db.DB.Model(&payment).Updates(map[string]interface{}{
				"status":  status,
				"amount":  amount,
				"paid_at": time.Now(),
			}).Error
		}
		return nil
	}

	// Créer un nouveau paiement
	payment = models.SubscriptionPayment{
		SubscriptionID:        subscriptionID,
		Amount:                amount,
		PaidAt:                time.Now(),
		StripePaymentIntentId: paymentIntentID,
		Status:                status,
	}
	return db.DB.Create(&payment).Error
}

func updateSubscriptionStatus(sub *models.Subscription) {
	newEnd := time.Now().AddDate(0, 1, 0)

	if sub.Status == models.SubscriptionPending {
		db.DB.Model(sub).Updates(map[string]interface{}{
			"status":   models.SubscriptionActive,
			"end_date": newEnd,
		})
	} else if sub.Status == models.SubscriptionActive {
		db.DB.Model(sub).Update("end_date", newEnd)
	}
}

func handlePaymentIntentCreated(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		utils.LogError(err, "Error parsing PaymentIntent created dans handlePaymentIntentCreated")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing PaymentIntent created"})
		return
	}

	utils.LogSuccess("PaymentIntent created dans handlePaymentIntentCreated")
	c.JSON(http.StatusOK, gin.H{"message": "PaymentIntent created - logged"})
}

func handlePaymentIntentProcessing(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		utils.LogError(err, "Error parsing PaymentIntent processing dans handlePaymentIntentProcessing")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing PaymentIntent processing"})
		return
	}

	utils.LogSuccess("PaymentIntent processing dans handlePaymentIntentProcessing")
	c.JSON(http.StatusOK, gin.H{"message": "PaymentIntent processing - logged"})
}

func handlePaymentIntentSucceeded(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		utils.LogError(err, "Error parsing PaymentIntent succeeded dans handlePaymentIntentSucceeded")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing PaymentIntent succeeded"})
		return
	}

	if pi.Customer == nil || pi.ID == "" {
		utils.LogError(nil, "PaymentIntent missing customer or ID dans handlePaymentIntentSucceeded")
		c.JSON(http.StatusBadRequest, gin.H{"error": "PaymentIntent missing customer or ID"})
		return
	}

	sub, err := findSubscriptionByCustomer(pi.Customer.ID, true)
	if err != nil {
		utils.LogError(err, "Subscription not found, will retry dans handlePaymentIntentSucceeded")
		c.JSON(http.StatusOK, gin.H{"message": "Subscription not ready, will retry"})
		return
	}

	if err := upsertSubscriptionPayment(sub.ID, int(pi.AmountReceived), pi.ID, models.SubscriptionPaymentSucceeded); err != nil {
		if err.Error() == "payment already recorded" {
			utils.LogError(err, "Payment already recorded dans handlePaymentIntentSucceeded")
			c.JSON(http.StatusOK, gin.H{"message": "Payment already recorded"})
			return
		}
		utils.LogError(err, "Error creating payment dans handlePaymentIntentSucceeded")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating payment"})
		return
	}

	utils.LogSuccess("Subscription activated via payment_intent.succeeded dans handlePaymentIntentSucceeded")
	updateSubscriptionStatus(sub)
	c.JSON(http.StatusOK, gin.H{"message": "Subscription activated via payment_intent.succeeded"})
}

func handlePaymentIntentFailed(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		utils.LogError(err, "Error parsing PaymentIntent failed dans handlePaymentIntentFailed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing PaymentIntent failed"})
		return
	}

	if pi.Customer == nil || pi.ID == "" {
		utils.LogError(nil, "PaymentIntent missing customer or ID dans handlePaymentIntentFailed")
		c.JSON(http.StatusOK, gin.H{"message": "PaymentIntent missing customer or ID"})
		return
	}

	sub, err := findSubscriptionByCustomer(pi.Customer.ID, true)
	if err != nil {
		utils.LogError(err, "Subscription not found, will retry dans handlePaymentIntentFailed")
		c.JSON(http.StatusOK, gin.H{"message": "Subscription not ready, will retry"})
		return
	}

	_ = upsertSubscriptionPayment(sub.ID, int(pi.Amount), pi.ID, models.SubscriptionPaymentFailed)

	utils.LogError(nil, "Payment failed dans handlePaymentIntentFailed pour subscription: "+sub.ID)

	if sub.Status == models.SubscriptionPending {
		db.DB.Model(sub).Update("status", models.SubscriptionCanceled)
	}

	utils.LogSuccess("Payment failed - subscription canceled if pending dans handlePaymentIntentFailed")
	c.JSON(http.StatusOK, gin.H{"message": "Payment failed - subscription canceled if pending"})
}

func handlePaymentIntentCanceled(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		utils.LogError(err, "Error parsing PaymentIntent canceled dans handlePaymentIntentCanceled")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing PaymentIntent canceled"})
		return
	}

	if pi.Customer == nil || pi.ID == "" {
		utils.LogError(nil, "PaymentIntent missing customer or ID dans handlePaymentIntentCanceled")
		c.JSON(http.StatusOK, gin.H{"message": "PaymentIntent missing customer or ID"})
		return
	}

	sub, err := findSubscriptionByCustomer(pi.Customer.ID, true)
	if err != nil {
		utils.LogError(err, "Subscription not found, will retry dans handlePaymentIntentCanceled")
		c.JSON(http.StatusOK, gin.H{"message": "Subscription not ready, will retry"})
		return
	}

	_ = upsertSubscriptionPayment(sub.ID, int(pi.Amount), pi.ID, models.SubscriptionPaymentCanceled)

	utils.LogError(nil, "Payment canceled dans handlePaymentIntentCanceled pour subscription: "+sub.ID)

	if sub.Status == models.SubscriptionPending {
		db.DB.Model(sub).Update("status", models.SubscriptionCanceled)
	}

	utils.LogSuccess("Payment canceled - subscription canceled if pending dans handlePaymentIntentCanceled")
	c.JSON(http.StatusOK, gin.H{"message": "Payment canceled - subscription canceled if pending"})
}

func handleInvoicePaymentSucceeded(c *gin.Context, event stripe.Event) {
	var invoiceData map[string]interface{}
	if err := json.Unmarshal(event.Data.Raw, &invoiceData); err != nil {
		utils.LogError(err, "Error parsing Invoice dans handleInvoicePaymentSucceeded")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing Invoice"})
		return
	}

	utils.LogSuccess("Received invoiceData dans handleInvoicePaymentSucceeded")

	var stripeSubID string
	if parent, ok := invoiceData["parent"].(map[string]interface{}); ok {
		if subDetails, ok := parent["subscription_details"].(map[string]interface{}); ok {
			if sub, ok := subDetails["subscription"].(string); ok && sub != "" {
				stripeSubID = sub
			}
		}
	}

	if stripeSubID == "" {
		utils.LogError(nil, "Impossible to retrieve subscription ID dans handleInvoicePaymentSucceeded")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	sub, err := findSubscriptionByStripeID(stripeSubID)
	if err != nil {
		utils.LogError(err, "Subscription not found, will retry dans handleInvoicePaymentSucceeded")
		c.JSON(http.StatusOK, gin.H{"message": "Subscription not ready, will retry"})
		return
	}

	var paymentIntentID string
	if pi, ok := invoiceData["payment_intent"].(string); ok {
		paymentIntentID = pi
	}

	var amount int
	if amountPaid, ok := invoiceData["amount_paid"].(float64); ok {
		amount = int(amountPaid)
	} else {
		utils.LogError(nil, "amount_paid missing or invalid dans handleInvoicePaymentSucceeded")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount"})
		return
	}

	if err := upsertSubscriptionPayment(sub.ID, amount, paymentIntentID, models.SubscriptionPaymentSucceeded); err != nil {
		if err.Error() == "payment already recorded" {
			utils.LogError(err, "Payment already recorded dans handleInvoicePaymentSucceeded")
			c.JSON(http.StatusOK, gin.H{"message": "Payment already recorded"})
			return
		}
		utils.LogError(err, "Error creating payment dans handleInvoicePaymentSucceeded")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating payment"})
		return
	}

	utils.LogSuccess("Subscription activated via invoice.payment_succeeded dans handleInvoicePaymentSucceeded")
	updateSubscriptionStatus(sub)

	var message string
	if sub.Status == models.SubscriptionPending {
		message = "Subscription activated via invoice.payment_succeeded"
	} else {
		message = "Subscription renewed via invoice.payment_succeeded"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
}

func handleInvoicePaymentFailed(c *gin.Context, event stripe.Event) {
	var invoiceData map[string]interface{}
	if err := json.Unmarshal(event.Data.Raw, &invoiceData); err != nil {
		utils.LogError(err, "Error parsing Invoice failed dans handleInvoicePaymentFailed")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing Invoice"})
		return
	}

	var stripeSubID string
	if v, ok := invoiceData["subscription"]; ok && v != nil {
		if s, ok := v.(string); ok && s != "" {
			stripeSubID = s
		}
	}

	if stripeSubID == "" {
		utils.LogError(nil, "Impossible to retrieve subscription ID for failed payment dans handleInvoicePaymentFailed")
		c.JSON(http.StatusOK, gin.H{"message": "Invalid subscription ID - event ignored"})
		return
	}

	var paymentIntentID string
	if pi, ok := invoiceData["payment_intent"].(string); ok {
		paymentIntentID = pi
	}

	sub, err := findSubscriptionByStripeID(stripeSubID)
	if err == nil {
		_ = upsertSubscriptionPayment(sub.ID, 0, paymentIntentID, models.SubscriptionPaymentFailed)
	}

	utils.LogError(nil, "Failed payment for subscription: "+stripeSubID+", PaymentIntent: "+paymentIntentID)

	c.JSON(http.StatusOK, gin.H{"message": "Invoice payment failed - logged"})
}
