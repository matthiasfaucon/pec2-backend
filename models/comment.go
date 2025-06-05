package models

import (
	"time"
)

type Comment struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PostID        string    `json:"postId" gorm:"column:post_id"`
	UserID        string    `json:"userId" gorm:"column:user_id"`
	Content       string    `json:"content" binding:"required"`
	CommentsCount int       `json:"commentsCount" gorm:"column:comments_count;default:0"`
	CreatedAt     time.Time `json:"createdAt"`
}

func (Comment) TableName() string {
	return "comments"
}
