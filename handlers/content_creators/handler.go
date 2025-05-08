package content_creators

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	mailsmodels "pec2-backend/utils/mails-models"

	"github.com/gin-gonic/gin"
)

// @Summary Apply to become a content creator
// @Description Submit an application to become a content creator
// @Tags content-creators
// @Accept multipart/form-data
// @Produce json
// @Param companyName formData string true "Company name" default(My Creative Company)
// @Param companyType formData string true "Company type" default(SARL)
// @Param siretNumber formData string true "SIRET number" default(12345678901234)
// @Param vatNumber formData string false "VAT number" default(FR12345678901)
// @Param streetAddress formData string true "Street address" default(123 Business Street)
// @Param postalCode formData string true "Postal code" default(75001)
// @Param city formData string true "City" default(Paris)
// @Param country formData string true "Country" default(France)
// @Param iban formData string true "IBAN" default(FR7630006000011234567890189)
// @Param bic formData string true "BIC" default(BNPAFRPP)
// @Param file formData file true "Document proof (PDF, image)"
// @Success 201 {object} map[string]interface{} "message: Application submitted successfully"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 409 {object} map[string]interface{} "error: Application already exists"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Security BearerAuth
// @Router /content-creators [post]
func Apply(c *gin.Context) {
	var contentCreatorInfoCreate models.ContentCreatorInfoCreate

	if err := c.ShouldBind(&contentCreatorInfoCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	userID := c.MustGet("user_id").(string)

	var user models.User
	if err := db.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching user information",
		})
		return
	}

	var existingApplication models.ContentCreatorInfo
	if err := db.DB.Where("user_id = ?", userID).First(&existingApplication).Error; err == nil {
		if existingApplication.Verified {
			c.JSON(http.StatusConflict, gin.H{
				"error": "You are already a content creator",
			})
		} else {
			c.JSON(http.StatusConflict, gin.H{
				"error": "You have already applied to become a content creator",
			})
		}
		return
	}

	isValid, err := utils.VerifySiret(contentCreatorInfoCreate.SiretNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error verifying SIRET number: " + err.Error(),
		})
		return
	}
	if !isValid {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid SIRET number: The provided SIRET number does not exist or is not active",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil || file == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Document proof is required",
		})
		return
	}

	documentURL, err := utils.UploadImage(file, "content_creator_documents", "document")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error uploading document: " + err.Error(),
		})
		return
	}

	contentCreatorInfo := models.ContentCreatorInfo{
		UserID:           userID,
		CompanyName:      contentCreatorInfoCreate.CompanyName,
		CompanyType:      contentCreatorInfoCreate.CompanyType,
		SiretNumber:      contentCreatorInfoCreate.SiretNumber,
		VatNumber:        contentCreatorInfoCreate.VatNumber,
		StreetAddress:    contentCreatorInfoCreate.StreetAddress,
		PostalCode:       contentCreatorInfoCreate.PostalCode,
		City:             contentCreatorInfoCreate.City,
		Country:          contentCreatorInfoCreate.Country,
		Iban:             contentCreatorInfoCreate.Iban,
		Bic:              contentCreatorInfoCreate.Bic,
		DocumentProofUrl: documentURL,
		Verified:         false,
	}

	result := db.DB.Create(&contentCreatorInfo)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	mailsmodels.ContentCreatorConfirmation(mailsmodels.ContentCreatorConfirmationData{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		CompanyName: contentCreatorInfo.CompanyName,
		CompanyType: contentCreatorInfo.CompanyType,
		SiretNumber: contentCreatorInfo.SiretNumber,
	})

	c.JSON(http.StatusCreated, gin.H{
		"message": "Content creator application submitted successfully",
	})
}
