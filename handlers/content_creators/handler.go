package content_creators

import (
	"fmt"
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
		if existingApplication.Status == models.ContentCreatorStatusApproved {
			c.JSON(http.StatusConflict, gin.H{
				"error": "You are already a content creator. Please use the update endpoint if you need to modify your information",
			})
			return
		} else if existingApplication.Status == models.ContentCreatorStatusPending {
			c.JSON(http.StatusConflict, gin.H{
				"error": "You have already applied to become a content creator",
			})
			return
		} else if existingApplication.Status == models.ContentCreatorStatusRejected {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Your application was rejected. Please use the update endpoint to resubmit your application",
			})
			return
		}
	}

	var existingSiret models.ContentCreatorInfo
	if err := db.DB.Where("siret_number = ? AND user_id != ?", contentCreatorInfoCreate.SiretNumber, userID).First(&existingSiret).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "This SIRET number is already registered by another content creator",
		})
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
		Status:           models.ContentCreatorStatusPending,
	}

	result := db.DB.Create(&contentCreatorInfo)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	mailsmodels.ContentCreatorConfirmation(mailsmodels.ContentCreatorConfirmationData{
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		CompanyName:   contentCreatorInfo.CompanyName,
		CompanyType:   contentCreatorInfo.CompanyType,
		SiretNumber:   contentCreatorInfo.SiretNumber,
		VatNumber:     contentCreatorInfo.VatNumber,
		StreetAddress: contentCreatorInfo.StreetAddress,
		PostalCode:    contentCreatorInfo.PostalCode,
		City:          contentCreatorInfo.City,
		Country:       contentCreatorInfo.Country,
		Iban:          contentCreatorInfo.Iban,
		Bic:           contentCreatorInfo.Bic,
	})

	utils.LogSuccessWithUser(userID, "Content creator application submitted successfully in Apply")
	c.JSON(http.StatusCreated, gin.H{
		"message": "Content creator application submitted successfully",
	})
}

// @Summary Update a content creator application
// @Description Update an existing content creator application (rejected or approved)
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
// @Success 200 {object} map[string]interface{} "message: Application updated successfully"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 404 {object} map[string]interface{} "error: No application found"
// @Failure 403 {object} map[string]interface{} "error: Application cannot be updated"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Security BearerAuth
// @Router /content-creators [put]
func UpdateContentCreatorInfo(c *gin.Context) {
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
	if err := db.DB.Where("user_id = ?", userID).First(&existingApplication).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No application found for this user",
		})
		return
	}

	if existingApplication.Status != models.ContentCreatorStatusRejected &&
		existingApplication.Status != models.ContentCreatorStatusApproved {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Application cannot be updated. Your application is currently pending",
		})
		return
	}

	if existingApplication.SiretNumber != contentCreatorInfoCreate.SiretNumber {
		var existingSiret models.ContentCreatorInfo
		if err := db.DB.Where("siret_number = ? AND user_id != ?", contentCreatorInfoCreate.SiretNumber, userID).First(&existingSiret).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{
				"error": "This SIRET number is already registered by another content creator",
			})
			return
		}
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

	oldDocumentURL := existingApplication.DocumentProofUrl

	documentURL, err := utils.UploadImage(file, "content_creator_documents", "document")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error uploading document: " + err.Error(),
		})
		return
	}

	if oldDocumentURL != "" {
		if err := utils.DeleteImage(oldDocumentURL); err != nil {
			fmt.Printf("Error deleting old document: %v\n", err)
		}
	}

	previousStatus := existingApplication.Status

	existingApplication.CompanyName = contentCreatorInfoCreate.CompanyName
	existingApplication.CompanyType = contentCreatorInfoCreate.CompanyType
	existingApplication.SiretNumber = contentCreatorInfoCreate.SiretNumber
	existingApplication.VatNumber = contentCreatorInfoCreate.VatNumber
	existingApplication.StreetAddress = contentCreatorInfoCreate.StreetAddress
	existingApplication.PostalCode = contentCreatorInfoCreate.PostalCode
	existingApplication.City = contentCreatorInfoCreate.City
	existingApplication.Country = contentCreatorInfoCreate.Country
	existingApplication.Iban = contentCreatorInfoCreate.Iban
	existingApplication.Bic = contentCreatorInfoCreate.Bic
	existingApplication.DocumentProofUrl = documentURL

	if previousStatus == models.ContentCreatorStatusRejected {
		existingApplication.Status = models.ContentCreatorStatusPending
	}

	if err := db.DB.Save(&existingApplication).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	mailsmodels.ContentCreatorUpdate(mailsmodels.ContentCreatorUpdateData{
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Email:         user.Email,
		CompanyName:   existingApplication.CompanyName,
		CompanyType:   existingApplication.CompanyType,
		SiretNumber:   existingApplication.SiretNumber,
		VatNumber:     existingApplication.VatNumber,
		StreetAddress: existingApplication.StreetAddress,
		PostalCode:    existingApplication.PostalCode,
		City:          existingApplication.City,
		Country:       existingApplication.Country,
		Iban:          existingApplication.Iban,
		Bic:           existingApplication.Bic,
	})

	utils.LogSuccessWithUser(userID, "Content creator information updated successfully in UpdateContentCreatorInfo")
	c.JSON(http.StatusOK, gin.H{
		"message": "Content creator information updated successfully",
	})
}

// @Summary Get all content creator applications (Admin)
// @Description Retrieves a list of all content creator applications (Admin access only)
// @Tags content-creators
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ContentCreatorInfo
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /content-creators/all [get]
func GetAllContentCreators(c *gin.Context) {
	var contentCreators []models.ContentCreatorInfo
	result := db.DB.Order("created_at DESC").Find(&contentCreators)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, contentCreators)
}

// @Summary Update content creator application status (Admin)
// @Description Update the status of a content creator application (Admin access only)
// @Tags content-creators
// @Accept json
// @Produce json
// @Param id path string true "Content Creator Application ID"
// @Param status body models.ContentCreatorStatusUpdate true "New status"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Status updated successfully"
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 404 {object} map[string]string "error: Content creator application not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /content-creators/{id}/status [put]
func UpdateContentCreatorStatus(c *gin.Context) {
	id := c.Param("id")
	var statusUpdate models.ContentCreatorStatusUpdate

	if err := c.ShouldBindJSON(&statusUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input data: " + err.Error(),
		})
		return
	}

	validStatus := false
	for _, status := range []models.ContentCreatorStatusType{
		models.ContentCreatorStatusPending,
		models.ContentCreatorStatusApproved,
		models.ContentCreatorStatusRejected,
	} {
		if statusUpdate.Status == status {
			validStatus = true
			break
		}
	}

	if !validStatus {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid status",
		})
		return
	}

	var contentCreator models.ContentCreatorInfo
	if result := db.DB.First(&contentCreator, "id = ?", id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Content creator application not found",
		})
		return
	}

	var user models.User
	if err := db.DB.First(&user, "id = ?", contentCreator.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error fetching user information",
		})
		return
	}

	if result := db.DB.Model(&contentCreator).Update("status", statusUpdate.Status); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	var newRole models.Role
	if statusUpdate.Status == models.ContentCreatorStatusApproved {
		newRole = models.ContentCreator
	} else {
		newRole = models.UserRole
	}

	if result := db.DB.Model(&user).Update("role", newRole); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error updating user role: " + result.Error.Error(),
		})
		return
	}

	mailsmodels.ContentCreatorStatusUpdate(mailsmodels.ContentCreatorStatusUpdateData{
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		CompanyName: contentCreator.CompanyName,
		Status:      statusUpdate.Status,
	})

	utils.LogSuccessWithUser(user.ID, "Content creator status updated successfully in UpdateContentCreatorStatus")
	c.JSON(http.StatusOK, gin.H{
		"message": "Status updated successfully",
	})
}
