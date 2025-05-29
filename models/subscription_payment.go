package models

import (
	"time"
)

type SubscriptionPayment struct {
	ID                    string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SubscriptionID        string    `json:"subscriptionId" gorm:"type:uuid;not null"`
	Amount                int       `json:"amount"`
	PaidAt                time.Time `json:"paidAt"`
	StripePaymentIntentId string    `json:"stripePaymentIntentId"`
	CreatedAt             time.Time `json:"createdAt"`
	UpdatedAt             time.Time `json:"updatedAt"`
}
