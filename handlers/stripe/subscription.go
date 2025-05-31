package stripe

import (
	"net/http"
	"os"

	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	stripe "github.com/stripe/stripe-go/v82"
	session "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
	stripeSubscription "github.com/stripe/stripe-go/v82/subscription"
)

// CreateSubscriptionCheckoutSession start a stripe payment to subscribe to a content creator (verified role). Returns the Stripe session ID to use on the frontend.
// @Summary Create a Stripe Checkout session for subscription
// @Description Start a Stripe payment to subscribe to a content creator (verified role). Returns the Stripe session ID to use on the frontend.
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param contentCreatorId path string true "ID of the content creator"
// @Security BearerAuth
// @Success 200 {object} map[string]string "sessionId: ID of the Stripe Checkout session, url: Stripe Checkout URL"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Can only subscribe to a content creator"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Stripe error or server error"
// @Router /subscriptions/checkout/{contentCreatorId} [post]
func CreateSubscriptionCheckoutSession(c *gin.Context) {
	contentCreatorId := c.Param("contentCreatorId")

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not authenticated dans CreateSubscriptionCheckoutSession")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var payer models.User
	err := db.DB.First(&payer, "id = ?", userID).Error
	if err != nil {
		utils.LogErrorWithUser(userID, err, "User not found dans CreateSubscriptionCheckoutSession")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var creator models.User
	err = db.DB.First(&creator, "id = ?", contentCreatorId).Error
	if err != nil {
		utils.LogErrorWithUser(userID, err, "Content creator not found dans CreateSubscriptionCheckoutSession")
		c.JSON(http.StatusNotFound, gin.H{"error": "Content creator not found"})
		return
	}
	if creator.Role != models.ContentCreator {
		utils.LogErrorWithUser(userID, nil, "Can only subscribe to a content creator dans CreateSubscriptionCheckoutSession")
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only subscribe to a content creator"})
		return
	}

	var existingSub models.Subscription
	err = db.DB.Where("user_id = ? AND content_creator_id = ? AND status IN (?)",
		payer.ID, creator.ID, []models.SubscriptionStatus{models.SubscriptionActive, models.SubscriptionPending}).First(&existingSub).Error
	if err == nil {
		utils.LogErrorWithUser(userID, nil, "Déjà une subscription active ou pending dans CreateSubscriptionCheckoutSession")
		c.JSON(http.StatusConflict, gin.H{"error": "You already have an active or pending subscription with this creator."})
		return
	}

	if payer.StripeCustomerId != "" {
		// Vérifie que le customer existe vraiment sur Stripe
		_, err := customer.Get(payer.StripeCustomerId, nil)
		if err != nil {
			// S'il n'existe pas, on le recrée
			payer.StripeCustomerId = ""
		}
	}
	if payer.StripeCustomerId == "" {
		custParams := &stripe.CustomerParams{
			Name: stripe.String(payer.UserName),
		}
		cust, err := customer.New(custParams)
		if err != nil {
			utils.LogErrorWithUser(userID, err, "Erreur lors de la création du client Stripe dans CreateSubscriptionCheckoutSession")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Erreur lors de la création du client Stripe"})
			return
		}
		db.DB.Model(&payer).Update("stripe_customer_id", cust.ID)
		payer.StripeCustomerId = cust.ID
	}

	params := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(payer.StripeCustomerId),
		PaymentMethodTypes: stripe.StringSlice([]string{"card"}),
		Mode:               stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String("price_1RUBlC4PRo6qYhfZsvXhuq8y"),
				Quantity: stripe.Int64(1),
			},
		},
		SuccessURL:        stripe.String("https://tonsite.com/success"),
		CancelURL:         stripe.String("https://tonsite.com/cancel"),
		ClientReferenceID: stripe.String(contentCreatorId),
	}

	s, err := session.New(params)
	if err != nil {
		utils.LogErrorWithUser(userID, err, "Erreur lors de la création de la session Stripe dans CreateSubscriptionCheckoutSession")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	utils.LogSuccessWithUser(userID, "Session Stripe de souscription créée avec succès dans CreateSubscriptionCheckoutSession")
	c.JSON(http.StatusOK, gin.H{"sessionId": s.ID, "url": s.URL})
}

// CancelSubscription cancels a Stripe subscription and updates its status in the database
// @Summary Cancel a subscription
// @Description Cancel a Stripe subscription and update its status in the database
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscriptionId path string true "ID of the subscription to cancel"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Subscription canceled successfully"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: You are not authorized to cancel this subscription"
// @Failure 404 {object} map[string]string "error: Subscription not found"
// @Failure 500 {object} map[string]string "error: Error when canceling the Stripe subscription"
// @Router /subscriptions/{subscriptionId} [delete]
func CancelSubscription(c *gin.Context) {
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")
	subscriptionId := c.Param("subscriptionId")

	// Validation de l'UUID
	if _, err := uuid.Parse(subscriptionId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not authenticated dans CancelSubscription")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var subscription models.Subscription
	err := db.DB.First(&subscription, "id = ?", subscriptionId).Error
	if err != nil {
		utils.LogErrorWithUser(userID, err, "Subscription not found dans CancelSubscription")
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if subscription.UserID != userID {
		utils.LogErrorWithUser(userID, nil, "Not authorized to cancel this subscription dans CancelSubscription")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to cancel this subscription"})
		return
	}

	_, err = stripeSubscription.Cancel(subscription.StripeSubscriptionId, &stripe.SubscriptionCancelParams{
		Prorate: stripe.Bool(false),
	})
	if err != nil {
		utils.LogErrorWithUser(userID, err, "Erreur lors de l'annulation Stripe dans CancelSubscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when canceling the Stripe subscription"})
		return
	}

	err = db.DB.Model(&subscription).Update("status", models.SubscriptionCanceled).Error
	if err != nil {
		utils.LogErrorWithUser(userID, err, "Erreur lors de la mise à jour du statut dans CancelSubscription")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when updating the subscription status"})
		return
	}

	utils.LogSuccessWithUser(userID, "Abonnement annulé avec succès dans CancelSubscription")
	c.JSON(http.StatusOK, gin.H{"message": "Subscription canceled successfully"})
}

// GetUserSubscriptions get all the subscriptions (active, canceled, history) of the connected user
// @Summary List the user's subscriptions
// @Description Return all the subscriptions (active, canceled, history) of the connected user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Subscription
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Router /subscriptions [get]
func GetUserSubscriptions(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not authenticated dans GetUserSubscriptions")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var subscriptions []models.Subscription
	err := db.DB.Where("user_id = ?", userID).Order("created_at DESC").Find(&subscriptions).Error
	if err != nil {
		utils.LogErrorWithUser(userID, err, "Erreur lors de la récupération des abonnements dans GetUserSubscriptions")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching subscriptions"})
		return
	}

	utils.LogSuccessWithUser(userID, "Liste des abonnements récupérée avec succès dans GetUserSubscriptions")
	c.JSON(http.StatusOK, subscriptions)
}

// GetSubscriptionDetail returns the details of a subscription
// @Summary Details of a subscription
// @Description Return the detailed information of a subscription
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscriptionId path string true "ID of the subscription"
// @Security BearerAuth
// @Success 200 {object} models.Subscription
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: You are not authorized to view this subscription"
// @Failure 404 {object} map[string]string "error: Subscription not found"
// @Router /subscriptions/{subscriptionId} [get]
func GetSubscriptionDetail(c *gin.Context) {
	subscriptionId := c.Param("subscriptionId")

	if _, err := uuid.Parse(subscriptionId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription ID"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not authenticated dans GetSubscriptionDetail")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var subscription models.Subscription
	err := db.DB.First(&subscription, "id = ?", subscriptionId).Error
	if err != nil {
		utils.LogErrorWithUser(userID, err, "Subscription not found dans GetSubscriptionDetail")
		c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		return
	}

	if subscription.UserID != userID {
		utils.LogErrorWithUser(userID, nil, "Not authorized to view this subscription dans GetSubscriptionDetail")
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not authorized to view this subscription"})
		return
	}

	utils.LogSuccessWithUser(userID, "Détail d'abonnement récupéré avec succès dans GetSubscriptionDetail")
	c.JSON(http.StatusOK, subscription)
}
