package content_creators

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"

	"github.com/gin-gonic/gin"
)

// @Summary Apply to become a content creator
// @Description Submit an application to become a content creator
// @Tags content-creators
// @Accept json
// @Produce json
// @Param contentCreatorInfo body models.ContentCreatorInfoCreate true "Content Creator Information"
// @Success 201 {object} map[string]interface{} "message: Application submitted successfully"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 409 {object} map[string]interface{} "error: Application already exists"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Security BearerAuth
// @Router /content-creators [post]
func Apply(c *gin.Context) {
	var contentCreatorInfoCreate models.ContentCreatorInfoCreate

	if err := c.ShouldBindJSON(&contentCreatorInfoCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// Get current user ID from context (set by JWTAuth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User ID not found in token",
		})
		return
	}

	// Check if user already has a content creator application
	var existingApplication models.ContentCreatorInfo
	if err := db.DB.Where("user_id = ?", userID).First(&existingApplication).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "You already have a content creator application",
		})
		return
	}

	// Create new content creator application
	contentCreatorInfo := models.ContentCreatorInfo{
		UserID:           userID.(string),
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
		DocumentProofUrl: contentCreatorInfoCreate.DocumentProofUrl,
		Verified:         false, // Default to false, needs admin verification
	}

	result := db.DB.Create(&contentCreatorInfo)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Content creator application submitted successfully",
	})
}
