package auth

import (
	"errors"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	mailsmodels "pec2-backend/utils/mails-models"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags auth
// @Accept json
// @Produce json
// @Param user body models.UserCreate true "User information"
// @Success 201 {object} map[string]interface{} "message: User created successfully, email: user email"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 409 {object} map[string]interface{} "error: Email already exists"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Router /register [post]
func CreateUser(c *gin.Context) {
	var userCreate models.UserCreate

	if err := c.ShouldBindJSON(&userCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	if userCreate.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The email cannot be empty",
		})
		return
	}

	if !utils.ValidateEmail(userCreate.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	if userCreate.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password cannot be empty",
		})
		return
	}

	if len(userCreate.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password must contain at least 6 characters",
		})
		return
	}

	hasLower := strings.ContainsAny(userCreate.Password, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(userCreate.Password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(userCreate.Password, "0123456789")

	if !hasLower || !hasUpper || !hasDigit {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password must contain at least one lowercase, one uppercase and one digit",
		})
		return
	}

	if userCreate.UserName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The username cannot be empty",
		})
		return
	}

	if userCreate.FirstName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The first name cannot be empty",
		})
		return
	}

	if userCreate.LastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The last name cannot be empty",
		})
		return
	}

	if userCreate.Sexe != models.Male && userCreate.Sexe != models.Female && userCreate.Sexe != models.Other {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The sexe must be MAN, WOMAN or OTHER",
		})
		return
	}

	if userCreate.BirthDayDate.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The birth date must be in the past",
		})
		return
	}

	var existingUser models.User
	if err := db.DB.Where("email = ?", userCreate.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "This email is already used",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error when checking the email existence",
		})
		return
	}

	if err := db.DB.Where("user_name = ?", userCreate.UserName).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "This username is already taken",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error when checking the username existence",
		})
		return
	}

	passwordHash, err := hashPassword(userCreate.Password)

	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Error when hashing password"})
		return
	}

	now := time.Now()
	code := utils.GenerateCode()

	user := models.User{
		Email:               userCreate.Email,
		Password:            passwordHash,
		UserName:            userCreate.UserName,
		FirstName:           userCreate.FirstName,
		LastName:            userCreate.LastName,
		BirthDayDate:        userCreate.BirthDayDate,
		Sexe:                userCreate.Sexe,
		Role:                models.UserRole,
		Bio:                 "",
		ProfilePicture:      "",
		StripeCustomerId:    "",
		SubscriptionPrice:   0,
		Enable:              true,
		SubscriptionEnable:  true,
		CommentsEnable:      true,
		MessageEnable:       true,
		EmailVerifiedAt:     nil,
		Siret:               "",
		ConfirmationCode:    code,
		ConfirmationCodeEnd: now.Add(1 * time.Hour),
	}

	result := db.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	resultSaveUser := db.DB.Save(&user)
	if resultSaveUser.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": resultSaveUser.Error.Error(),
		})
		return
	}

	mailsmodels.ConfirmEmail(user.Email, code)
	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"email":   user.Email,
	})
}

// LoginRequest model for login
// @Description model for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"utilisateur@exemple.com"`
	Password string `json:"password" binding:"required,min=6" example:"Motdepasse123"`
}

// @Summary user login
// @Description user login with credential
// @Tags auth
// @Accept json
// @Produce json
// @Param user body LoginRequest true "User credentials"
// @Success 200 {object} map[string]interface{} "token: JWT token"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 401 {object} map[string]interface{} "error: Wrong credentials or email not verified"
// @Failure 422 {object} map[string]interface{} "error: JWT not generated"
// @Router /login [post]
func Login(c *gin.Context) {
	var inputLogin LoginRequest

	if err := c.ShouldBindJSON(&inputLogin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	if inputLogin.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The email cannot be empty",
		})
		return
	}

	if !utils.ValidateEmail(inputLogin.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	if inputLogin.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password cannot be empty",
		})
		return
	}

	var user models.User
	result := db.DB.Where("email = ?", inputLogin.Email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error: " + result.Error.Error(),
			})
		}
		return
	}

	isSamePassword := samePassword(inputLogin.Password, user.Password)

	if !isSamePassword {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Wrong credentials",
		})
		return
	}

	if user.EmailVerifiedAt == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user don't valid email",
		})
		return
	}

	token, err := utils.GenerateJWT(user, 72)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Error when generating JWT"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
	})
}

// @Summary Validation email
// @Description After create account, user valid it email
// @Tags auth
// @Accept json
// @Produce json
// @Param code path string true "user code received by mail"
// @Success 200 {object} map[string]interface{} "message": "User validate account"
// @Failure 400 {object} map[string]interface{} "error: User already validated account"
// @Router /valid-email/{token} [get]
func ValidEmail(c *gin.Context) {
	code := c.Param("code")
	var user models.User

	result := db.DB.Where("confirmation_code = ?", code).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error: " + result.Error.Error(),
			})
		}
		return
	}

	if time.Now().After(user.ConfirmationCodeEnd) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Confirmation code expired",
		})
		return
	}

	if user.EmailVerifiedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User already validated account",
		})
		return
	}

	now := time.Now()
	user.EmailVerifiedAt = &now
	user.ConfirmationCode = ""

	resultSaveUser := db.DB.Save(&user)
	if resultSaveUser.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": resultSaveUser.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User validate account",
	})
}

// @Summary Resend Validation email
// @Description Resend validation email for users who loose their code or code is expired
// @Tags auth
// @Accept json
// @Produce json
// @Param email path string true "user send email for received new mail"
// @Success 200 {object} map[string]interface{} "message": "send email at user email address"
// @Failure 400 {object} map[string]interface{} "error: User already validated account"
// @Failure 404 {object} map[string]interface{} "error: User not found"
// @Router /valid-email/{token} [get]
func ResendValidEmail(c *gin.Context) {
	email := c.Param("email")

	var user models.User
	result := db.DB.Where("email = ?", email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error: " + result.Error.Error(),
			})
		}
		return
	}

	if user.EmailVerifiedAt != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User already validated account",
		})
		return
	}

	now := time.Now()
	code := utils.GenerateCode()

	user.ConfirmationCode = code
	user.ConfirmationCodeEnd = now.Add(1 * time.Hour)

	if result := db.DB.Save(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user: " + result.Error.Error()})
		return
	}

	mailsmodels.ConfirmEmail(email, code)
	c.JSON(http.StatusOK, gin.H{
		"message": "Resend code for user : " + user.ID,
	})

}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func samePassword(formPassword string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(formPassword))
	return err == nil
}
