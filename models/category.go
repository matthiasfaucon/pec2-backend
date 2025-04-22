package models

import (
	"time"
)

type Category struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Name      string     `json:"name" binding:"required"`
	Posts     []Post     `json:"posts,omitempty" gorm:"many2many:post_categories;"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

type CategoryCreate struct {
	Name string `json:"name" binding:"required"`
}

func (Category) TableName() string {
	return "categories"
}

