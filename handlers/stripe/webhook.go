package stripe

import (
	"encoding/json"
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
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Impossible de lire le corps de la requête"})
		return
	}

	secret := os.Getenv("STRIPE_WEBHOOK_SECRET")
	if secret == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook secret non configuré"})
		return
	}

	sig := c.GetHeader("Stripe-Signature")
	event, err := webhook.ConstructEvent(payload, sig, secret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Vérification de la signature Stripe échouée"})
		return
	}

	switch event.Type {
	case "checkout.session.completed":
		handleCheckoutSessionCompleted(c, event)
	case "payment_intent.succeeded":
		handlePaymentIntentSucceeded(c, event)
	case "payment_intent.failed":
		handlePaymentIntentFailed(c, event)
	case "invoice.payment_succeeded":
		handleInvoicePaymentSucceeded(c, event)
	default:
		c.JSON(http.StatusOK, gin.H{"message": "Événement ignoré"})
	}
}

func handleCheckoutSessionCompleted(c *gin.Context, event stripe.Event) {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing CheckoutSession"})
		return
	}

	if session.Customer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer manquant dans la session"})
		return
	}

	customerID := session.Customer.ID
	creatorID := session.ClientReferenceID
	if creatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ClientReferenceID manquant"})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "stripe_customer_id = ?", customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé pour ce customer"})
		return
	}

	var creator models.User
	if err := db.DB.First(&creator, "id = ?", creatorID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Créateur non trouvé"})
		return
	}

	if creator.Role != models.ContentCreator {
		c.JSON(http.StatusForbidden, gin.H{"error": "La cible n'est pas un content creator"})
		return
	}

	var stripeSubID string
	if session.Subscription != nil {
		stripeSubID = session.Subscription.ID
		var tmp models.Subscription
		if err := db.DB.First(&tmp, "stripe_subscription_id = ?", stripeSubID).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Subscription Stripe déjà existante"})
			return
		}
	}

	var dup models.Subscription
	if err := db.DB.Where("user_id = ? AND content_creator_id = ? AND status IN ?",
		user.ID, creator.ID,
		[]models.SubscriptionStatus{models.SubscriptionPending, models.SubscriptionActive}).
		First(&dup).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription locale déjà existante"})
		return
	}

	now := time.Now()
	end := now.AddDate(0, 1, 0)
	
	// Déterminer le statut initial selon le statut de paiement
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création subscription"})
		return
	}

	if initialStatus == models.SubscriptionActive {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription créée et activée"})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription créée, en attente du paiement"})
	}
}

func handlePaymentIntentSucceeded(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing PaymentIntent réussi"})
		return
	}

	if pi.Customer == nil || pi.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "PaymentIntent sans customer ou ID"})
		return
	}

	customerID := pi.Customer.ID

	var user models.User
	if err := db.DB.First(&user, "stripe_customer_id = ?", customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// SEUL ENDROIT qui active une subscription Pending
	var sub models.Subscription
	if err := db.DB.
		Where("user_id = ? AND status = ? AND created_at > ?", user.ID, models.SubscriptionPending, time.Now().Add(-1*time.Hour)).
		Order("created_at desc").
		First(&sub).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription correspondante introuvable"})
		return
	}

	var existing models.SubscriptionPayment
	if err := db.DB.First(&existing, "stripe_payment_intent_id = ?", pi.ID).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Paiement déjà enregistré"})
		return
	}

	payment := models.SubscriptionPayment{
		SubscriptionID:        sub.ID,
		Amount:                int(pi.AmountReceived),
		PaidAt:                time.Now(),
		StripePaymentIntentId: pi.ID,
	}

	if err := db.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création du paiement"})
		return
	}

	// ACTIVATION: Pending -> Active
	newEnd := time.Now().AddDate(0, 1, 0)
	db.DB.Model(&sub).Updates(map[string]interface{}{
		"status":   models.SubscriptionActive,
		"end_date": newEnd,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Subscription activée via payment_intent.succeeded"})
}

func handleInvoicePaymentSucceeded(c *gin.Context, event stripe.Event) {
	var invoiceData map[string]interface{}
	if err := json.Unmarshal(event.Data.Raw, &invoiceData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing Invoice"})
		return
	}

	subscription, ok := invoiceData["subscription"]
	if !ok || subscription == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice sans subscription"})
		return
	}

	stripeSubID, ok := subscription.(string)
	if !ok || stripeSubID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID invalide"})
		return
	}

	var sub models.Subscription
	if err := db.DB.First(&sub, "stripe_subscription_id = ?", stripeSubID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription locale non trouvée"})
		return
	}

	// Pour les renouvellements: seulement étendre la date, pas changer le statut
	// (la subscription doit déjà être Active)
	if sub.Status != models.SubscriptionActive {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tentative de renouvellement d'une subscription non active"})
		return
	}

	// Enregistrer le paiement de renouvellement
	if amountPaid, ok := invoiceData["amount_paid"].(float64); ok {
		payment := models.SubscriptionPayment{
			SubscriptionID: sub.ID,
			Amount:         int(amountPaid),
			PaidAt:         time.Now(),
			// Note: invoice n'a pas forcément de payment_intent_id
		}
		db.DB.Create(&payment)
	}

	// RENOUVELLEMENT: étendre seulement la date de fin
	newEnd := time.Now().AddDate(0, 1, 0)
	db.DB.Model(&sub).Update("end_date", newEnd)

	c.JSON(http.StatusOK, gin.H{"message": "Subscription renouvelée via invoice.payment_succeeded"})
}

func handlePaymentIntentFailed(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing PaymentIntent échoué"})
		return
	}

	if pi.Customer == nil {
		c.JSON(http.StatusOK, gin.H{"message": "PaymentIntent échoué sans customer"})
		return
	}

	customerID := pi.Customer.ID
	var user models.User
	if err := db.DB.First(&user, "stripe_customer_id = ?", customerID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Utilisateur non trouvé pour l'échec de paiement"})
		return
	}

	// SEUL ENDROIT qui annule une subscription Pending après échec de paiement
	var sub models.Subscription
	if err := db.DB.
		Where("user_id = ? AND status = ? AND created_at > ?", user.ID, models.SubscriptionPending, time.Now().Add(-1*time.Hour)).
		Order("created_at desc").
		First(&sub).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Aucune subscription pending à annuler"})
		return
	}

	// ANNULATION: Pending -> Canceled
	db.DB.Model(&sub).Update("status", models.SubscriptionCanceled)

	c.JSON(http.StatusOK, gin.H{"message": "Subscription annulée suite à l'échec de paiement"})
}