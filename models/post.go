package models

import (
	"time"
)

type Post struct {
	ID         string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID     string     `json:"userId" gorm:"column:user_id"`
	Name       string     `json:"name" binding:"required"`
	PictureURL string     `json:"pictureUrl" gorm:"column:picture_url"`
	IsFree     bool       `json:"isFree" gorm:"default:true"`
	Enable     bool       `json:"enable" gorm:"default:true"`
	Categories []Category `json:"categories" gorm:"many2many:post_categories;"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	DeletedAt  *time.Time `json:"deletedAt,omitempty" gorm:"index"`
}

type PostCreate struct {
	Name       string   `json:"name" binding:"required"`
	IsFree     bool     `json:"isFree"`
	PictureURL string   `json:"pictureUrl"`
	Categories []string `json:"categories"`
}

type PostUpdate struct {
	Name       string   `json:"name"`
	IsFree     bool     `json:"isFree"`
	Categories []string `json:"categories"`
	Enable     bool    `json:"enable"`
}

func (Post) TableName() string {
	return "posts"
}

