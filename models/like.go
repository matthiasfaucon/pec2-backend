package models

import (
	"time"
)

type Like struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID    string    `json:"postId" gorm:"column:post_id"`
	UserID    string    `json:"userId" gorm:"column:user_id"`
	CreatedAt time.Time `json:"createdAt"`
}
