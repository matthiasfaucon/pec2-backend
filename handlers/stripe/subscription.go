package stripe

import (
	"net/http"
	"os"

	"pec2-backend/db"
	"pec2-backend/models"

	"github.com/gin-gonic/gin"
	stripe "github.com/stripe/stripe-go/v82"
	session "github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/customer"
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

	var user models.User
	err := db.DB.First(&user, "id = ?", contentCreatorId).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	if user.Role != models.ContentCreator {
		c.JSON(http.StatusForbidden, gin.H{"error": "Can only subscribe to a content creator"})
		return
	}

	if user.StripeCustomerId == "" {
		custParams := &stripe.CustomerParams{
			Email: stripe.String(user.Email),
			Name:  stripe.String(user.UserName),
		}
		cust, err := customer.New(custParams)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error when creating Stripe customer"})
			return
		}
		db.DB.Model(&user).Update("stripe_customer_id", cust.ID)
		user.StripeCustomerId = cust.ID
	}

	params := &stripe.CheckoutSessionParams{
		Customer:           stripe.String(user.StripeCustomerId),
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessionId": s.ID, "url": s.URL})
}
