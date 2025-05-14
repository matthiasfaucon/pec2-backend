package models

import (
	"time"
)

// PrivateMessage represents a message sent between two users
type PrivateMessage struct {
	ID         string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SenderID   string     `json:"senderId" gorm:"column:sender_id"`
	ReceiverID string     `json:"receiverId" gorm:"column:receiver_id"`
	Content    string     `json:"content" binding:"required"`
	Status     string     `json:"status" gorm:"default:UNREAD"` // UNREAD, READ
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty" gorm:"index"`
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
