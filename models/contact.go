package models

import (
	"time"

	"gorm.io/gorm"
)

type Contact struct {
	gorm.Model
	ID          string    `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	FirstName   string    `json:"firstName" gorm:"column:first_name" binding:"required"`
	LastName    string    `json:"lastName" gorm:"column:last_name" binding:"required"`
	Email       string    `json:"email" binding:"required,email"`
	Subject     string    `json:"subject" binding:"required"`
	Message     string    `json:"message" gorm:"type:text" binding:"required"`
	SubmittedAt time.Time `json:"submittedAt" gorm:"column:submitted_at;default:CURRENT_TIMESTAMP"`
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
