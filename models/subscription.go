package models

import (
	"time"
)

type SubscriptionStatus string

const (
	SubscriptionActive   SubscriptionStatus = "ACTIVE"
	SubscriptionCanceled SubscriptionStatus = "CANCELED"
	SubscriptionPending  SubscriptionStatus = "PENDING"
)

type Subscription struct {
	ID                   string             `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID               string             `json:"userId" gorm:"type:uuid;not null"`
	ContentCreatorID     string             `json:"contentCreatorId" gorm:"type:uuid;not null"`
	Status               SubscriptionStatus `json:"status" gorm:"type:varchar(20);default:'PENDING'"`
	StripeSubscriptionId string             `json:"stripeSubscriptionId"`
	StartDate            time.Time          `json:"startDate"`
	EndDate              *time.Time         `json:"endDate"`
	CreatedAt            time.Time          `json:"createdAt"`
	UpdatedAt            time.Time          `json:"updatedAt"`
}
