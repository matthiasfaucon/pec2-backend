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

// Fonction commune pour créer un paiement
func createSubscriptionPayment(subscriptionID string, amount int, paymentIntentID string) error {
	// Vérification si le paiement existe déjà
	if paymentIntentID != "" {
		var existing models.SubscriptionPayment
		if err := db.DB.First(&existing, "stripe_payment_intent_id = ?", paymentIntentID).Error; err == nil {
			return fmt.Errorf("paiement déjà enregistré")
		}
	}

	// Création du paiement
	payment := models.SubscriptionPayment{
		SubscriptionID:        subscriptionID,
		Amount:                amount,
		PaidAt:                time.Now(),
		StripePaymentIntentId: paymentIntentID,
	}

	return db.DB.Create(&payment).Error
}

// Fonction commune pour activer/renouveler une subscription
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

	var sub models.Subscription
	if err := db.DB.
		Where("user_id = ? AND status = ? AND created_at > ?", user.ID, models.SubscriptionPending, time.Now().Add(-1*time.Hour)).
		Order("created_at desc").
		First(&sub).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription correspondante introuvable"})
		return
	}

	// Utilisation de la fonction commune
	if err := createSubscriptionPayment(sub.ID, int(pi.AmountReceived), pi.ID); err != nil {
		if err.Error() == "paiement déjà enregistré" {
			c.JSON(http.StatusOK, gin.H{"message": "Paiement déjà enregistré"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création du paiement"})
		return
	}

	updateSubscriptionStatus(&sub)
	c.JSON(http.StatusOK, gin.H{"message": "Subscription activée via payment_intent.succeeded"})
}

func handleInvoicePaymentSucceeded(c *gin.Context, event stripe.Event) {
	var invoiceData map[string]interface{}
	if err := json.Unmarshal(event.Data.Raw, &invoiceData); err != nil {
		fmt.Printf("[DEBUG] Erreur parsing Invoice: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Erreur parsing Invoice"})
		return
	}

	fmt.Printf("[DEBUG] invoiceData reçu: %+v\n", invoiceData)

	// Récupération de l'ID subscription
	var stripeSubID string
	if parent, ok := invoiceData["parent"].(map[string]interface{}); ok {
		if subDetails, ok := parent["subscription_details"].(map[string]interface{}); ok {
			if sub, ok := subDetails["subscription"].(string); ok && sub != "" {
				stripeSubID = sub
			}
		}
	}

	if stripeSubID == "" {
		if v, ok := invoiceData["subscription"]; ok && v != nil {
			if s, ok := v.(string); ok && s != "" {
				stripeSubID = s
			}
		}
	}

	if stripeSubID == "" {
		fmt.Println("[DEBUG] Impossible de récupérer l'ID de subscription")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Subscription ID invalide"})
		return
	}

	var sub models.Subscription
	if err := db.DB.First(&sub, "stripe_subscription_id = ?", stripeSubID).Error; err != nil {
		fmt.Printf("[DEBUG] Subscription locale non trouvée pour %s: %v\n", stripeSubID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription locale non trouvée"})
		return
	}

	// Récupération du payment_intent ID
	var paymentIntentID string
	if pi, ok := invoiceData["payment_intent"].(string); ok {
		paymentIntentID = pi
	}

	// Récupération du montant
	var amount int
	if amountPaid, ok := invoiceData["amount_paid"].(float64); ok {
		amount = int(amountPaid)
	} else {
		fmt.Println("[DEBUG] amount_paid absent ou invalide")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Montant invalide"})
		return
	}

	// Utilisation de la fonction commune
	if err := createSubscriptionPayment(sub.ID, amount, paymentIntentID); err != nil {
		if err.Error() == "paiement déjà enregistré" {
			fmt.Printf("[DEBUG] Paiement déjà enregistré pour PI: %s\n", paymentIntentID)
			c.JSON(http.StatusOK, gin.H{"message": "Paiement déjà enregistré"})
			return
		}
		fmt.Printf("[DEBUG] Erreur création paiement: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur création du paiement"})
		return
	}

	updateSubscriptionStatus(&sub)

	var message string
	if sub.Status == models.SubscriptionPending {
		message = "Subscription activée via invoice.payment_succeeded"
	} else {
		message = "Subscription renouvelée via invoice.payment_succeeded"
	}

	c.JSON(http.StatusOK, gin.H{"message": message})
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

	var sub models.Subscription
	if err := db.DB.
		Where("user_id = ? AND status = ? AND created_at > ?", user.ID, models.SubscriptionPending, time.Now().Add(-1*time.Hour)).
		Order("created_at desc").
		First(&sub).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Aucune subscription pending à annuler"})
		return
	}

	db.DB.Model(&sub).Update("status", models.SubscriptionCanceled)

	c.JSON(http.StatusOK, gin.H{"message": "Subscription annulée suite à l'échec de paiement"})
}
