package models

import (
	"time"
)

// ContentCreatorInfo represents a content creator application
type ContentCreatorInfo struct {
	ID               string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID           string    `json:"userId" gorm:"type:uuid;not null"`
	CompanyName      string    `json:"companyName" binding:"required"`
	CompanyType      string    `json:"companyType" binding:"required"`
	SiretNumber      string    `json:"siretNumber" binding:"required"`
	VatNumber        string    `json:"vatNumber"`
	StreetAddress    string    `json:"streetAddress" binding:"required"`
	PostalCode       string    `json:"postalCode" binding:"required"`
	City             string    `json:"city" binding:"required"`
	Country          string    `json:"country" binding:"required"`
	Iban             string    `json:"iban" binding:"required"`
	Bic              string    `json:"bic" binding:"required"`
	DocumentProofUrl string    `json:"documentProofUrl" binding:"required"`
	Verified         bool      `json:"verified" gorm:"default:false"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

func (ContentCreatorInfo) TableName() string {
	return "content_creator_info"
}

// ContentCreatorInfoCreate model for creating content creator info
// @Description model for applying to become a content creator
type ContentCreatorInfoCreate struct {
	CompanyName      string `json:"companyName" binding:"required" example:"Creative Studios"`
	CompanyType      string `json:"companyType" binding:"required" example:"Limited Company"`
	SiretNumber      string `json:"siretNumber" binding:"required" example:"12345678901234"`
	VatNumber        string `json:"vatNumber" example:"FR12345678901"`
	StreetAddress    string `json:"streetAddress" binding:"required" example:"123 Content Street"`
	PostalCode       string `json:"postalCode" binding:"required" example:"75001"`
	City             string `json:"city" binding:"required" example:"Paris"`
	Country          string `json:"country" binding:"required" example:"France"`
	Iban             string `json:"iban" binding:"required" example:"FR7630006000011234567890189"`
	Bic              string `json:"bic" binding:"required" example:"BNPAFRPP"`
	DocumentProofUrl string `json:"documentProofUrl" binding:"required" example:"https://storage.example.com/documents/proof123.pdf"`
}
