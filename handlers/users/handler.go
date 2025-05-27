package users

import (
	"errors"
	"fmt"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	mailsmodels "pec2-backend/utils/mails-models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Struct pour Swagger : demande de code de réinitialisation
// @Description Email pour demander un code de réinitialisation
// @name PasswordResetRequest
// @Param email body string true "Email de l'utilisateur"
type PasswordResetRequest struct {
	Email string `json:"email" example:"utilisateur@exemple.com"`
}

// Struct pour Swagger : confirmation de réinitialisation
// @Description Email, code et nouveau mot de passe pour confirmer la réinitialisation
// @name PasswordResetConfirm
// @Param email body string true "Email de l'utilisateur"
// @Param code body string true "Code reçu par email"
// @Param newPassword body string true "Nouveau mot de passe"
type PasswordResetConfirm struct {
	Email       string `json:"email" example:"utilisateur@exemple.com"`
	Code        string `json:"code" example:"123456"`
	NewPassword string `json:"newPassword" example:"NouveauMotdepasse123"`
}

// @Summary Get all users (Admin)
// @Description Retrieves a list of all users (Admin access only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "users: array of user objects"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden - Admin access required"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /users [get]
func GetAllUsers(c *gin.Context) {
	var users []models.User
	result := db.DB.Order("created_at DESC").Find(&users)

	if result.Error != nil {
		utils.LogError(result.Error, "Error when retrieving all users in GetAllUsers")
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	for i := range users {
		users[i].Password = ""
	}

	utils.LogSuccess("List of users retrieved successfully in GetAllUsers")
	c.JSON(http.StatusOK, users)
}

// @Summary Update user password
// @Description Update user's password by verifying the old password and setting a new one
// @Tags users
// @Accept json
// @Produce json
// @Param password body models.PasswordUpdate true "Password update information"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Password updated successfully"
// @Failure 400 {object} map[string]string "error: Invalid request"
// @Failure 401 {object} map[string]string "error: Invalid old password"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Error updating password"
// @Router /users/password [put]
func UpdatePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(errors.New("user_id manquant"), "User not found dans UpdatePassword")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var passwordUpdate models.PasswordUpdate
	if err := c.ShouldBindJSON(&passwordUpdate); err != nil {
		utils.LogError(err, "Error when binding JSON in UpdatePassword")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data: " + err.Error()})
		return
	}

	if len(passwordUpdate.NewPassword) < 6 {
		utils.LogError(errors.New("new password is too short"), "New password is too short in UpdatePassword")
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must contain at least 6 characters"})
		return
	}

	if passwordUpdate.OldPassword == passwordUpdate.NewPassword {
		utils.LogError(errors.New("new password is the same as the old password"), "New password is the same as the old password in UpdatePassword")
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must be different from the old password"})
		return
	}

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		utils.LogError(result.Error, "User not found in UpdatePassword")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordUpdate.OldPassword)); err != nil {
		utils.LogError(err, "Incorrect old password in UpdatePassword")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordUpdate.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.LogError(err, "Error when hashing the new password in UpdatePassword")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	if result := db.DB.Model(&user).Update("password", string(hashedPassword)); result.Error != nil {
		utils.LogError(result.Error, "Error when updating password in UpdatePassword")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
		return
	}

	utils.LogSuccess("Password updated successfully in UpdatePassword")
	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// @Summary Update user profile
// @Description Update the current authenticated user's profile information with optional profile picture
// @Tags users
// @Accept multipart/form-data
// @Produce json
// @Param userName formData string false "UserName"
// @Param firstName formData string false "First name"
// @Param lastName formData string false "Last name"
// @Param bio formData string false "Biography"
// @Param email formData string false "Email address"
// @Param sexe formData string false "Sexe"
// @Param birthDayDate formData string false "BirthDayDate"
// @Param profilePicture formData file false "Profile picture image file"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "message: Profile updated successfully, user: updated user object"
// @Failure 400 {object} map[string]string "error: Invalid request data"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Error updating profile"
// @Router /users/profile [put]
func UpdateUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(errors.New("user_id manquant"), "User not found in token dans UpdateUserProfile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		utils.LogError(result.Error, "User not found in UpdateUserProfile")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var formData models.UserUpdateFormData
	if err := c.ShouldBind(&formData); err != nil {
		utils.LogError(err, "Error when binding form data in UpdateUserProfile")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data: " + err.Error()})
		return
	}

	if formData.UserName != "" {
		var existingUser models.User
		if err := db.DB.Where("user_name = ? AND id != ?", formData.UserName, userID).First(&existingUser).Error; err == nil {
			utils.LogError(errors.New("username déjà utilisé"), "Username already taken in UpdateUserProfile")
			c.JSON(http.StatusConflict, gin.H{
				"error": "This username is already taken",
			})
			return
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			utils.LogError(err, "Error when checking the username existence in UpdateUserProfile")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error when checking the username existence",
			})
			return
		}
		user.UserName = formData.UserName
	}
	if formData.Bio != "" {
		user.Bio = formData.Bio
	}
	if formData.Email != "" {
		if !utils.ValidateEmail(formData.Email) {
			utils.LogError(errors.New("format email invalide"), "Invalid email format in UpdateUserProfile")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
			return
		}
		user.Email = formData.Email
	}

	if formData.FirstName != "" {
		user.FirstName = formData.FirstName
	}
	if formData.LastName != "" {
		user.LastName = formData.LastName
	}

	file, err := c.FormFile("profilePicture")
	if err == nil && file != nil {
		oldImageURL := user.ProfilePicture

		imageURL, err := utils.UploadImage(file, "profile_pictures", "profile")
		if err != nil {
			utils.LogError(err, "Error when uploading profile picture in UpdateUserProfile")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading profile picture: " + err.Error()})
			return
		}

		user.ProfilePicture = imageURL

		if oldImageURL != "" {
			_ = utils.DeleteImage(oldImageURL)
		}
	}

	if result := db.DB.Save(&user); result.Error != nil {
		utils.LogError(result.Error, "Error when saving user profile in UpdateUserProfile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile: " + result.Error.Error()})
		return
	}

	user.Password = ""

	utils.LogSuccess("User profile updated successfully in UpdateUserProfile")
	c.JSON(http.StatusOK, user)
}

// @Summary Get user profile
// @Description Get the current authenticated user's profile information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "user: user object"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: Error retrieving profile"
// @Router /users/profile [get]
func GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(errors.New("user_id manquant"), "User not found in token dans GetUserProfile")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var user models.User
	result := db.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		utils.LogError(result.Error, "User not found in GetUserProfile")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	user.Password = ""

	utils.LogSuccess("User profile retrieved successfully in GetUserProfile")
	c.JSON(http.StatusOK, user)
}

// UserStatsResponse représente la réponse de statistiques d'utilisateurs
type UserStatsResponse struct {
	Period string `json:"period"`
	Count  int    `json:"count"`
	Label  string `json:"label"`
}

// @Summary Get user statistics (Admin)
// @Description Get count of users by month or year
// @Tags users
// @Accept json
// @Produce json
// @Param filter query string true "Filter type: 'month' or 'year'"
// @Param year query int false "Year to filter by (default is current year)"
// @Param month query int false "Month to filter by (1-12, only used with 'month' filter)"
// @Security BearerAuth
// @Success 200 {array} UserStatsResponse
// @Failure 400 {object} map[string]string "error: Invalid filter parameter"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error retrieving statistics"
// @Router /users/statistics [get]
func GetUserStatistics(c *gin.Context) {
	filter := c.Query("filter")
	if filter != "month" && filter != "year" {
		utils.LogError(errors.New("paramètre filter invalide"), "Invalid filter parameter in GetUserStatistics")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter must be 'month' or 'year'"})
		return
	}

	currentYear := time.Now().Year()
	yearStr := c.DefaultQuery("year", strconv.Itoa(currentYear))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		utils.LogError(err, "Invalid year parameter in GetUserStatistics")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
		return
	}

	var stats []UserStatsResponse

	if filter == "month" {
		monthStr := c.Query("month")
		if monthStr != "" {
			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				utils.LogError(errors.New("paramètre month invalide"), "Invalid month parameter in GetUserStatistics")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month parameter (must be 1-12)"})
				return
			}

			stats, err = getUserCountByDay(year, month)
			if err != nil {
				utils.LogError(err, "Error when retrieving daily statistics in GetUserStatistics")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving daily statistics: " + err.Error()})
				return
			}
		} else {
			stats, err = getUserCountByMonth(year)
			if err != nil {
				utils.LogError(err, "Error when retrieving monthly statistics in GetUserStatistics")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving monthly statistics: " + err.Error()})
				return
			}
		}
	} else {
		stats, err = getUserCountByYear(year)
		if err != nil {
			utils.LogError(err, "Error when retrieving yearly statistics in GetUserStatistics")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving yearly statistics: " + err.Error()})
			return
		}
	}

	utils.LogSuccess("User statistics retrieved successfully in GetUserStatistics")
	c.JSON(http.StatusOK, stats)
}

func getUserCountByDay(year, month int) ([]UserStatsResponse, error) {
	daysInMonth := 31
	if month == 4 || month == 6 || month == 9 || month == 11 {
		daysInMonth = 30
	} else if month == 2 {
		if year%4 == 0 && (year%100 != 0 || year%400 == 0) {
			daysInMonth = 29
		} else {
			daysInMonth = 28
		}
	}

	var results []UserStatsResponse

	for day := 1; day <= daysInMonth; day++ {
		currentDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		nextDate := currentDate.AddDate(0, 0, 1)

		if currentDate.After(time.Now()) {
			continue
		}

		var count int64
		err := db.DB.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", currentDate, nextDate).
			Count(&count).Error

		if err != nil {
			return nil, err
		}

		dayStr := fmt.Sprintf("%02d", day)
		monthStr := fmt.Sprintf("%02d", month)

		results = append(results, UserStatsResponse{
			Period: fmt.Sprintf("%d-%s-%s", year, monthStr, dayStr),
			Count:  int(count),
			Label:  fmt.Sprintf("%d %s", day, currentDate.Format("Jan")),
		})
	}

	return results, nil
}

func getUserCountByMonth(year int) ([]UserStatsResponse, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	if startDate.After(time.Now()) {
		return []UserStatsResponse{}, nil
	}

	var results []UserStatsResponse

	for month := 1; month <= 12; month++ {
		currentDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		nextDate := currentDate.AddDate(0, 1, 0)

		if currentDate.After(time.Now()) {
			continue
		}

		var count int64
		err := db.DB.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", currentDate, nextDate).
			Count(&count).Error

		if err != nil {
			return nil, err
		}

		monthName := currentDate.Format("Jan")
		results = append(results, UserStatsResponse{
			Period: fmt.Sprintf("%d-%02d", year, month),
			Count:  int(count),
			Label:  monthName,
		})
	}

	return results, nil
}

func getUserCountByYear(targetYear int) ([]UserStatsResponse, error) {
	var oldestUser models.User
	err := db.DB.Order("created_at ASC").First(&oldestUser).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []UserStatsResponse{}, nil
		}
		return nil, err
	}

	startYear := oldestUser.CreatedAt.Year()
	currentYear := time.Now().Year()

	if targetYear > 0 {
		if targetYear > currentYear {
			return []UserStatsResponse{}, nil
		}
		if targetYear < startYear {
			return []UserStatsResponse{}, nil
		}

		startDate := time.Date(targetYear, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(targetYear+1, 1, 1, 0, 0, 0, 0, time.UTC)

		var count int64
		err := db.DB.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", startDate, endDate).
			Count(&count).Error

		if err != nil {
			return nil, err
		}

		return []UserStatsResponse{
			{
				Period: fmt.Sprintf("%d", targetYear),
				Count:  int(count),
				Label:  fmt.Sprintf("%d", targetYear),
			},
		}, nil
	}

	var results []UserStatsResponse

	for year := startYear; year <= currentYear; year++ {
		startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)

		var count int64
		err := db.DB.Model(&models.User{}).
			Where("created_at >= ? AND created_at < ?", startDate, endDate).
			Count(&count).Error

		if err != nil {
			return nil, err
		}

		results = append(results, UserStatsResponse{
			Period: fmt.Sprintf("%d", year),
			Count:  int(count),
			Label:  fmt.Sprintf("%d", year),
		})
	}

	return results, nil
}

// @Summary Get user role statistics (Admin)
// @Description Get the count of users for each role (Admin access only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int "Role counts"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/stats/roles [get]
func GetUserRoleStats(c *gin.Context) {
	var roleCounts = make(map[string]int)

	roleCounts["ADMIN"] = 0
	roleCounts["CONTENT_CREATOR"] = 0
	roleCounts["USER"] = 0

	for _, role := range []models.Role{models.AdminRole, models.ContentCreator, models.UserRole} {
		var count int64
		if err := db.DB.Model(&models.User{}).Where("role = ?", role).Count(&count).Error; err != nil {
			utils.LogError(err, "Error when counting users by role in GetUserRoleStats")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users by role"})
			return
		}
		roleCounts[string(role)] = int(count)
	}

	utils.LogSuccess("User role statistics retrieved successfully in GetUserRoleStats")
	c.JSON(http.StatusOK, roleCounts)
}

// @Summary Get user gender statistics (Admin)
// @Description Get the count of users for each gender (Admin access only)
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]int "Gender counts"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Internal server error"
// @Router /users/stats/gender [get]
func GetUserGenderStats(c *gin.Context) {
	var genderCounts = make(map[string]int)

	genderCounts["MAN"] = 0
	genderCounts["WOMAN"] = 0
	genderCounts["OTHER"] = 0

	for _, sexe := range []models.Sexe{models.Male, models.Female, models.Other} {
		var count int64
		if err := db.DB.Model(&models.User{}).Where("sexe = ?", sexe).Count(&count).Error; err != nil {
			utils.LogError(err, "Error when counting users by gender in GetUserGenderStats")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error counting users by gender"})
			return
		}
		genderCounts[string(sexe)] = int(count)
	}

	utils.LogSuccess("User gender statistics retrieved successfully in GetUserGenderStats")
	c.JSON(http.StatusOK, genderCounts)
}

// @Summary send a reset password code
// @Description send a reset password code to the email if the user exists
// @Tags users
// @Accept json
// @Produce json
// @Param data body PasswordResetRequest true "Email of the user"
// @Success 200 {object} map[string]string "message: Code sent"
// @Failure 404 {object} map[string]string "error: User not found"
// @Router /users/password/reset/request [post]
func RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError(err, "Error when binding JSON in RequestPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email"})
		return
	}

	if !utils.ValidateEmail(req.Email) {
		utils.LogError(errors.New("format email invalide"), "Invalid email format in RequestPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.LogError(err, "User not found in RequestPasswordReset")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	code := utils.GenerateCode()
	end := time.Now().Add(15 * time.Minute)

	user.ResetPasswordCode = code
	user.ResetPasswordCodeEnd = end

	if err := db.DB.Save(&user).Error; err != nil {
		utils.LogError(err, "Error when saving the reset code in RequestPasswordReset")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving the user"})
		return
	}

	mailsmodels.SendResetPasswordCode(user.Email, code)
	utils.LogSuccess("Reset password code sent successfully in RequestPasswordReset")
	c.JSON(http.StatusOK, gin.H{"message": "Code sent to the email if it exists in our database."})
}

// @Summary Reset password with a code
// @Description Change the password if the code is correct and not expired
// @Tags users
// @Accept json
// @Produce json
// @Param data body PasswordResetConfirm true "Email, code, new password"
// @Success 200 {object} map[string]string "message: Password reset"
// @Failure 400 {object} map[string]string "error: Invalid data or code incorrect/expired"
// @Router /users/password/reset/confirm [post]
func ConfirmPasswordReset(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		Code        string `json:"code" binding:"required"`
		NewPassword string `json:"newPassword" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.LogError(err, "Error when binding JSON in ConfirmPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data"})
		return
	}

	if !utils.ValidateEmail(req.Email) {
		utils.LogError(errors.New("format email invalide"), "Invalid email format in ConfirmPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		utils.LogError(err, "User not found in ConfirmPasswordReset")
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.ResetPasswordCode != req.Code || time.Now().After(user.ResetPasswordCodeEnd) {
		utils.LogError(errors.New("code invalide ou expiré"), "Invalid or expired reset code in ConfirmPasswordReset")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid code or expired"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.LogError(err, "Error when hashing the new password in ConfirmPasswordReset")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing the password"})
		return
	}

	user.Password = string(hashedPassword)
	user.ResetPasswordCode = ""
	user.ResetPasswordCodeEnd = time.Time{}

	if err := db.DB.Save(&user).Error; err != nil {
		utils.LogError(err, "Error when saving the new password in ConfirmPasswordReset")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error saving the user"})
		return
	}

	utils.LogSuccess("Password reset successfully in ConfirmPasswordReset")
	c.JSON(http.StatusOK, gin.H{"message": "Password reset"})
}
