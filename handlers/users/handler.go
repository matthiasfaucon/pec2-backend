package users

import (
	"errors"
	"fmt"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	for i := range users {
		users[i].Password = ""
	}

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var passwordUpdate models.PasswordUpdate
	if err := c.ShouldBindJSON(&passwordUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data: " + err.Error()})
		return
	}

	if len(passwordUpdate.NewPassword) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must contain at least 6 characters"})
		return
	}

	// Vérifie si le nouveau mot de passe est identique à l'ancien
	if passwordUpdate.OldPassword == passwordUpdate.NewPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "The new password must be different from the old password"})
		return
	}

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordUpdate.OldPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Incorrect old password"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(passwordUpdate.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	if result := db.DB.Model(&user).Update("password", string(hashedPassword)); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating password"})
		return
	}

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var user models.User
	if result := db.DB.Where("id = ?", userID).First(&user); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var formData models.UserUpdateFormData
	if err := c.ShouldBind(&formData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid form data: " + err.Error()})
		return
	}

	if formData.UserName != "" {
		user.UserName = formData.UserName
	}
	if formData.Bio != "" {
		user.Bio = formData.Bio
	}
	if formData.Email != "" {
		if !utils.ValidateEmail(formData.Email) {
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading profile picture: " + err.Error()})
			return
		}

		user.ProfilePicture = imageURL

		if oldImageURL != "" {
			_ = utils.DeleteImage(oldImageURL)
		}
	}

	if result := db.DB.Save(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating profile: " + result.Error.Error()})
		return
	}

	user.Password = ""

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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var user models.User
	// Utiliser exactement la même approche que GetAllUsers qui fonctionne dans les tests
	result := db.DB.First(&user, "id = ?", userID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Ne pas renvoyer le mot de passe
	user.Password = ""

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter must be 'month' or 'year'"})
		return
	}

	currentYear := time.Now().Year()
	yearStr := c.DefaultQuery("year", strconv.Itoa(currentYear))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid year parameter"})
		return
	}

	var stats []UserStatsResponse

	if filter == "month" {
		monthStr := c.Query("month")
		if monthStr != "" {
			month, err := strconv.Atoi(monthStr)
			if err != nil || month < 1 || month > 12 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month parameter (must be 1-12)"})
				return
			}

			stats, err = getUserCountByDay(year, month)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving daily statistics: " + err.Error()})
				return
			}
		} else {
			stats, err = getUserCountByMonth(year)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving monthly statistics: " + err.Error()})
				return
			}
		}
	} else {
		stats, err = getUserCountByYear(year)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving yearly statistics: " + err.Error()})
			return
		}
	}

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
