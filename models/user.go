package models

import (
	"database/sql"
	"time"
)

// Définition des rôles utilisateur
type Role string

// Valeurs possibles pour le type Role
const (
	AdminRole  Role = "ADMIN"
	UserRole   Role = "USER"
	Subscriber Role = "SUBSCRIBER"
)

// User représente un utilisateur dans la base de données
type User struct {
	ID                     string       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Email                  string       `json:"email" binding:"required,email"`
	Password               string       `json:"password" binding:"required,min=6"`
	UserName               string       `json:"username"`
	Role                   Role         `json:"role"`
	Bio                    string       `json:"bio"`
	ProfilePicture         string       `json:"profilePicture"`
	StripeCustomerId       string       `json:"stripeCustomerId"`
	SubscriptionPrice      int          `json:"subscriptionPrice"`
	Enable                 bool         `json:"enable"`
	SubscriptionEnable     bool         `json:"subscriptionEnable"`
	CommentsEnable         bool         `json:"commentsEnable"`
	MessageEnable          bool         `json:"messageEnable"`
	EmailVerifiedAt        sql.NullTime `json:"emailVerifiedAt"`
	Siret                  string       `json:"siret"`
	TokenVerificationEmail string       `json:"TokenVerificationEmail"`
	CreatedAt              time.Time    `json:"createdAt"`
	UpdatedAt              time.Time    `json:"updatedAt"`
	DeletedAt              *time.Time   `json:"deletedAt,omitempty" gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

// UserCreate model for create a user
// @Description model for create a user
type UserCreate struct {
	Email    string `json:"email" binding:"required,email" example:"utilisateur@exemple.com"`
	Password string `json:"password" binding:"required,min=6" example:"Motdepasse123"`
}

// PasswordUpdate modèle pour la mise à jour du mot de passe
// @Description modèle pour mettre à jour le mot de passe d'un utilisateur
type PasswordUpdate struct {
	OldPassword string `json:"oldPassword" binding:"required" example:"AncienMotdepasse123"`
	NewPassword string `json:"newPassword" binding:"required,min=6" example:"NouveauMotdepasse123"`
}
