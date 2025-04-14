package models

import (
	"database/sql"

	"gorm.io/gorm"
)

// User représente un utilisateur dans la base de données

type Role string

// Définir les valeurs possibles pour le type Status
const (
	AdminRole  Role = "ADMIN"
	UserRole   Role = "USER"
	Subscriber Role = "SUBSCRIBER"
)

type User struct {
	gorm.Model
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
