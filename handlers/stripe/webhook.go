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

// invoicePayload sert à extraire manuellement les champs JSON de l’événement invoice.paid
type invoicePayload struct {
	SubscriptionID  string `json:"subscription"`
	PaymentIntentID string `json:"payment_intent"`
	AmountPaid      int64  `json:"amount_paid"`
}

// StripeWebhookHandler gère tous les webhooks Stripe
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
	case "payment_intent.payment_failed":
		handlePaymentIntentFailed(c, event)
	case "invoice.paid":
		handleInvoicePaid(c, event)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer missing in session"})
		return
	}
	customerID := session.Customer.ID
	creatorID := session.ClientReferenceID
	if creatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ClientReferenceID (créateur) manquant"})
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

	// Si Stripe Subscription déjà créée, on ignore
	var stripeSubID string
	if session.Subscription != nil {
		stripeSubID = session.Subscription.ID
		var tmp models.Subscription
		if err := db.DB.First(&tmp, "stripe_subscription_id = ?", stripeSubID).Error; err == nil {
			c.JSON(http.StatusOK, gin.H{"message": "Subscription Stripe déjà existante"})
			return
		}
	}

	// Éviter doublon local
	var dup models.Subscription
	if err := db.DB.
		Where("user_id = ? AND content_creator_id = ? AND status IN ?", user.ID, creator.ID,
			[]models.SubscriptionStatus{models.SubscriptionPending, models.SubscriptionActive}).
		First(&dup).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription locale déjà existante"})
		return
	}

	now := time.Now()
	end := now.AddDate(0, 1, 0)
	sub := models.Subscription{
		UserID:               user.ID,
		ContentCreatorID:     creator.ID,
		Status:               models.SubscriptionPending,
		StripeSubscriptionId: stripeSubID,
		StartDate:            now,
		EndDate:              &end,
	}
	if err := db.DB.Create(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création subscription"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription créée, en attente du paiement"})
}

func handlePaymentIntentSucceeded(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing PaymentIntent"})
		return
	}

	// Éviter doublons
	if err := db.DB.First(new(models.SubscriptionPayment), "stripe_payment_intent_id = ?", pi.ID).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Paiement déjà enregistré"})
		return
	}

	if pi.Customer == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Customer missing on PaymentIntent"})
		return
	}
	customerID := pi.Customer.ID

	var user models.User
	if err := db.DB.First(&user, "stripe_customer_id = ?", customerID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé pour ce customer"})
		return
	}

	var sub models.Subscription
	if err := db.DB.
		Where("user_id = ? AND status IN ?", user.ID,
			[]models.SubscriptionStatus{models.SubscriptionPending, models.SubscriptionActive}).
		Order("created_at DESC").
		First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription introuvable pour ce paiement"})
		return
	}

	payment := models.SubscriptionPayment{
		SubscriptionID:        sub.ID,
		Amount:                int(pi.Amount),
		PaidAt:                time.Now(),
		StripePaymentIntentId: pi.ID,
	}
	if err := db.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création du paiement"})
		return
	}

	if sub.EndDate != nil {
		newEnd := sub.EndDate.AddDate(0, 1, 0)
		db.DB.Model(&sub).Updates(map[string]interface{}{
			"end_date": newEnd,
			"status":   models.SubscriptionActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Paiement enregistré et subscription activée"})
}

func handlePaymentIntentFailed(c *gin.Context, event stripe.Event) {
	var pi stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &pi); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing PaymentIntent échoué"})
		return
	}

	if pi.Customer != nil {
		customerID := pi.Customer.ID
		var user models.User
		if err := db.DB.First(&user, "stripe_customer_id = ?", customerID).Error; err == nil {
			db.DB.
				Model(&models.Subscription{}).
				Where("user_id = ? AND status = ? AND created_at > ?", user.ID, models.SubscriptionPending, time.Now().Add(-1*time.Hour)).
				Update("status", models.SubscriptionCanceled)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Échec de paiement traité"})
}

func handleInvoicePaid(c *gin.Context, event stripe.Event) {
	// On parse manuellement uniquement les champs JSON nécessaires
	var inv invoicePayload
	if err := json.Unmarshal(event.Data.Raw, &inv); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing Invoice payload"})
		return
	}

	if inv.SubscriptionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice sans subscription"})
		return
	}
	if inv.PaymentIntentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice sans payment_intent"})
		return
	}

	var sub models.Subscription
	if err := db.DB.First(&sub, "stripe_subscription_id = ?", inv.SubscriptionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription introuvable pour cette invoice"})
		return
	}

	// Éviter doublon de paiement
	if err := db.DB.First(new(models.SubscriptionPayment), "stripe_payment_intent_id = ?", inv.PaymentIntentID).Error; err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Paiement déjà enregistré"})
		return
	}

	payment := models.SubscriptionPayment{
		SubscriptionID:        sub.ID,
		Amount:                int(inv.AmountPaid),
		PaidAt:                time.Now(),
		StripePaymentIntentId: inv.PaymentIntentID,
	}
	if err := db.DB.Create(&payment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création du paiement"})
		return
	}

	if sub.EndDate != nil {
		newEnd := sub.EndDate.AddDate(0, 1, 0)
		db.DB.Model(&sub).Updates(map[string]interface{}{
			"end_date": newEnd,
			"status":   models.SubscriptionActive,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Paiement invoice enregistré et subscription activée"})
}
