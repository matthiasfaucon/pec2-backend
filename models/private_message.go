package models

import (
	"time"
)

// MessageStatusType représente le type de statut d'un message
type MessageStatusType string

const (
	MessageStatusUnread   MessageStatusType = "UNREAD"
	MessageStatusRead     MessageStatusType = "READ"
	MessageStatusArchived MessageStatusType = "ARCHIVED"
	MessageStatusDeleted  MessageStatusType = "DELETED"
)

// PrivateMessage represents a message sent between two users
type PrivateMessage struct {
	ID         string            `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SenderID   string            `json:"senderId" gorm:"column:sender_id"`
	ReceiverID string            `json:"receiverId" gorm:"column:receiver_id"`
	Content    string            `json:"content" binding:"required"`
	Status     MessageStatusType `json:"status" gorm:"default:UNREAD"`
	CreatedAt  time.Time         `json:"createdAt"`
	UpdatedAt  time.Time         `json:"updatedAt"`
	DeletedAt  *time.Time        `json:"deletedAt,omitempty" gorm:"index"`
}

// PrivateMessageCreate model for creating a private message
// @Description model for creating a private message
type PrivateMessageCreate struct {
	ReceiverID string `json:"receiverId" binding:"required"`
	Content    string `json:"content" binding:"required"`
}

func (PrivateMessage) TableName() string {
	return "private_messages"
}
