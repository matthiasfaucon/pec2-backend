package report

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"

	"slices"

	"github.com/gin-gonic/gin"
)

// @Summary Report a post
// @Description Report a post for inappropriate content
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param report body models.ReportCreate true "Report reason"
// @Security BearerAuth
// @Success 201 {object} models.Report
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id}/report [post]
func ReportPost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in ReportPost")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	postID := c.Param("id")

	// Vérifier si le post existe
	var post models.Post
	if err := db.DB.First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in ReportPost")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}
	var reportCreate models.ReportCreate
	if err := c.ShouldBindJSON(&reportCreate); err != nil {
		utils.LogError(err, "Invalid input in ReportPost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// Vérifier que la raison est valide
	isValidReason := false
	validReasons := []models.ReportReason{
		models.DISLIKE, models.HARASSMENT, models.SELF_HARM,
		models.VIOLENCE, models.RESTRICTED_ITEMS, models.NUDITY,
		models.SCAM, models.MISINFORMATION, models.ILLEGAL_CONTENT,
	}

	if slices.Contains(validReasons, models.ReportReason(reportCreate.Reason)) {
		isValidReason = true
	}

	if !isValidReason {
		utils.LogError(nil, "Invalid report reason in ReportPost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report reason"})
		return
	}

	// Vérifier si l'utilisateur a déjà signalé ce post
	var existingReport models.Report
	if err := db.DB.Where("post_id = ? AND reported_by = ?", postID, userID).First(&existingReport).Error; err == nil {
		utils.LogError(nil, "Already reported in ReportPost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "You have already reported this post"})
		return
	}
	report := models.Report{
		PostID:     postID,
		ReportedBy: userID.(string),
		Reason:     models.ReportReason(reportCreate.Reason),
	}

	if err := db.DB.Create(&report).Error; err != nil {
		utils.LogError(err, "Error creating report in ReportPost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating report: " + err.Error()})
		return
	}

	utils.LogSuccessWithUser(userID, "Report successfully created in ReportPost")
	c.JSON(http.StatusCreated, report)
}

// @Summary Get all reports (Admin only)
// @Description Get all reports with optional filtering
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Report
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/reports [get]
func GetAllReports(c *gin.Context) {
	var reports []models.Report

	if err := db.DB.Order("created_at DESC").Find(&reports).Error; err != nil {
		utils.LogError(err, "Error retrieving reports in GetAllReports")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving reports: " + err.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Reports successfully retrieved in GetAllReports")
	c.JSON(http.StatusOK, reports)
}
