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
		utils.LogError(err, "Error when binding JSON in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	if userCreate.Email == "" {
		utils.LogError(errors.New("email vide"), "Email vide in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The email cannot be empty",
		})
		return
	}

	if !utils.ValidateEmail(userCreate.Email) {
		utils.LogError(errors.New("format email invalide"), "Invalid email format in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	if userCreate.Password == "" {
		utils.LogError(errors.New("mot de passe vide"), "Password empty in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password cannot be empty",
		})
		return
	}

	if len(userCreate.Password) < 6 {
		utils.LogError(errors.New("mot de passe trop court"), "Password too short in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password must contain at least 6 characters",
		})
		return
	}

	hasLower := strings.ContainsAny(userCreate.Password, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(userCreate.Password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(userCreate.Password, "0123456789")

	if !hasLower || !hasUpper || !hasDigit {
		utils.LogError(errors.New("mot de passe faible"), "Weak password in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password must contain at least one lowercase, one uppercase and one digit",
		})
		return
	}

	if userCreate.UserName == "" {
		utils.LogError(errors.New("username vide"), "Username empty in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The username cannot be empty",
		})
		return
	}

	if userCreate.FirstName == "" {
		utils.LogError(errors.New("firstname vide"), "FirstName empty in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The first name cannot be empty",
		})
		return
	}

	if userCreate.LastName == "" {
		utils.LogError(errors.New("lastname vide"), "LastName empty in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The last name cannot be empty",
		})
		return
	}

	if userCreate.Sexe != models.Male && userCreate.Sexe != models.Female && userCreate.Sexe != models.Other {
		utils.LogError(errors.New("sexe invalide"), "Invalid sexe in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The sexe must be MAN, WOMAN or OTHER",
		})
		return
	}

	if userCreate.BirthDayDate.After(time.Now()) {
		utils.LogError(errors.New("date de naissance future"), "Future birth date in CreateUser")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The birth date must be in the past",
		})
		return
	}

	var existingUser models.User
	if err := db.DB.Where("email = ?", userCreate.Email).First(&existingUser).Error; err == nil {
		utils.LogError(errors.New("email déjà utilisé"), "Email already used in CreateUser")
		c.JSON(http.StatusConflict, gin.H{
			"error": "This email is already used",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.LogError(err, "Error when checking the email existence in CreateUser")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error when checking the email existence",
		})
		return
	}

	if err := db.DB.Where("user_name = ?", userCreate.UserName).First(&existingUser).Error; err == nil {
		utils.LogError(errors.New("username déjà utilisé"), "Username already used in CreateUser")
		c.JSON(http.StatusConflict, gin.H{
			"error": "This username is already taken",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		utils.LogError(err, "Error when checking the username existence in CreateUser")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error when checking the username existence",
		})
		return
	}

	passwordHash, err := hashPassword(userCreate.Password)

	if err != nil {
		utils.LogError(err, "Error when hashing password in CreateUser")
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
		utils.LogError(result.Error, "Error when creating user in CreateUser")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	resultSaveUser := db.DB.Save(&user)
	if resultSaveUser.Error != nil {
		utils.LogError(resultSaveUser.Error, "Error when saving user in CreateUser")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": resultSaveUser.Error.Error(),
		})
		return
	}

	mailsmodels.ConfirmEmail(user.Email, code)
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "User created successfully in CreateUser")
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
		utils.LogError(err, "Error when binding JSON in Login")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	if inputLogin.Email == "" {
		utils.LogError(errors.New("email vide"), "Email empty in Login")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The email cannot be empty",
		})
		return
	}

	if !utils.ValidateEmail(inputLogin.Email) {
		utils.LogError(errors.New("format email invalide"), "Invalid email format in Login")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	if inputLogin.Password == "" {
		utils.LogError(errors.New("mot de passe vide"), "Password empty in Login")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password cannot be empty",
		})
		return
	}

	var user models.User
	result := db.DB.Where("email = ?", inputLogin.Email).First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			utils.LogError(result.Error, "User not found in Login")
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User not found",
			})
		} else {
			utils.LogError(result.Error, "Database error in Login")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error: " + result.Error.Error(),
			})
		}
		return
	}

	isSamePassword := samePassword(inputLogin.Password, user.Password)

	if !isSamePassword {
		utils.LogError(errors.New("mauvais mot de passe"), "Wrong password in Login")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Wrong credentials",
		})
		return
	}

	if user.EmailVerifiedAt == nil {
		utils.LogError(errors.New("email non vérifié"), "Email not verified in Login")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user don't valid email",
		})
		return
	}

	token, err := utils.GenerateJWT(user, 72)
	if err != nil {
		utils.LogError(err, "Error when generating JWT in Login")
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "Error when generating JWT"})
		return
	}

	userID := user.ID
	if userID == "" {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "User login successfully in Login")
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
			utils.LogError(result.Error, "User not found in ValidEmail")
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			utils.LogError(result.Error, "Database error in ValidEmail")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error: " + result.Error.Error(),
			})
		}
		return
	}

	if time.Now().After(user.ConfirmationCodeEnd) {
		utils.LogError(errors.New("code expiré"), "Confirmation code expired in ValidEmail")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Confirmation code expired",
		})
		return
	}

	if user.EmailVerifiedAt != nil {
		utils.LogError(errors.New("déjà validé"), "User already validated in ValidEmail")
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
		utils.LogError(resultSaveUser.Error, "Error when saving user in ValidEmail")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": resultSaveUser.Error.Error(),
		})
		return
	}

	userID := user.ID
	if userID == "" {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Email validation successfully in ValidEmail")
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
			utils.LogError(result.Error, "User not found in ResendValidEmail")
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			utils.LogError(result.Error, "Database error in ResendValidEmail")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Database error: " + result.Error.Error(),
			})
		}
		return
	}

	if user.EmailVerifiedAt != nil {
		utils.LogError(errors.New("déjà validé"), "User already validated in ResendValidEmail")
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
		utils.LogError(result.Error, "Error when saving user in ResendValidEmail")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating user: " + result.Error.Error()})
		return
	}

	mailsmodels.ConfirmEmail(email, code)
	userID := user.ID
	if userID == "" {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Resend validation code successfully in ResendValidEmail")
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
