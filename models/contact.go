package models

import (
	"time"
)

// StatusType définit les différents statuts possibles pour une demande de contact
type StatusType string

const (
	StatusOpen       StatusType = "open"
	StatusProcessing StatusType = "processing"
	StatusClosed     StatusType = "closed"
	StatusRejected   StatusType = "rejected"
)

// GormModel définit les champs communs pour Swagger
// @Description Champs communs du modèle Gorm
type GormModel struct {
	ID        uint       `json:"id" gorm:"primarykey"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty" gorm:"index"`
}

// Contact représente une demande de contact dans la base de données
// @Description Modèle complet d'une demande de contact
// @Scheme Contact
type Contact struct {
	ID          string     `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	FirstName   string     `json:"firstName" gorm:"column:first_name" binding:"required"`
	LastName    string     `json:"lastName" gorm:"column:last_name" binding:"required"`
	Email       string     `json:"email" binding:"required,email"`
	Subject     string     `json:"subject" binding:"required"`
	Message     string     `json:"message" gorm:"type:text" binding:"required"`
	Status      StatusType `json:"status" gorm:"type:varchar(20);default:'open'"`
	SubmittedAt time.Time  `json:"submittedAt" gorm:"column:submitted_at;default:CURRENT_TIMESTAMP"`
	CreatedAt   time.Time  `json:"createdAt" swaggerignore:"true"`
	UpdatedAt   time.Time  `json:"updatedAt" swaggerignore:"true"`
	DeletedAt   *time.Time `json:"deletedAt,omitempty" swaggerignore:"true" gorm:"index"`
}

func (Contact) TableName() string {
	return "contacts"
}

// ContactCreate modèle pour créer une demande de contact
// @Description modèle pour créer une demande de contact
type ContactCreate struct {
	FirstName string `json:"firstName" binding:"required" example:"Jean"`
	LastName  string `json:"lastName" binding:"required" example:"Dupont"`
	Email     string `json:"email" binding:"required,email" example:"jean.dupont@exemple.com"`
	Subject   string `json:"subject" binding:"required" example:"Demande d'information"`
	Message   string `json:"message" binding:"required" example:"J'aimerais avoir plus d'informations sur vos services."`
}

// ContactStatusUpdate modèle pour mettre à jour le statut d'une demande de contact
// @Description modèle pour mettre à jour le statut d'une demande de contact
type ContactStatusUpdate struct {
	Status StatusType `json:"status" binding:"required" example:"processing"`
}
