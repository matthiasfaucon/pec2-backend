package models

import (
	"time"
)

type Post struct {
	ID         string     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID     string     `json:"userId" gorm:"column:user_id"`
	Name       string     `json:"name" binding:"required"`
	PictureURL string     `json:"pictureUrl" gorm:"column:picture_url"`
	IsFree     bool       `json:"isFree" gorm:"default:false"`
	Enable     bool       `json:"enable" gorm:"default:true"`
	Categories []Category `json:"categories" gorm:"many2many:post_categories;"`
	Likes      []Like     `json:"likes,omitempty"`
	Reports    []Report   `json:"reports,omitempty"`
	User       User       `json:"user,omitempty" gorm:"foreignKey:UserID"`
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
	Enable     bool     `json:"enable"`
}

type PostResponse struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	PictureURL    string     `json:"pictureUrl"`
	IsFree        bool       `json:"isFree"`
	Enable        bool       `json:"enable"`
	Categories    []Category `json:"categories"`
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
	User          UserInfo   `json:"user"`
	LikesCount    int        `json:"likesCount"`
	CommentsCount int        `json:"commentsCount"`
	ReportsCount  int        `json:"reportsCount"`
}

type UserInfo struct {
	ID             string `json:"id"`
	UserName       string `json:"userName"`
	ProfilePicture string `json:"profilePicture"`
}

func (Post) TableName() string {
	return "posts"
}
