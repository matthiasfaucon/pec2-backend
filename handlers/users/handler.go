package users

import (
	"database/sql"
	"errors"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.UserCreate true "User information"
// @Success 200 {object} map[string]interface{} "message: User created successfully, email: user email"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 409 {object} map[string]interface{} "error: Email already exists"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Router /user [post]
func CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + err.Error(),
		})
		return
	}

	// 1. Validation de l'email
	if user.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The email cannot be empty",
		})
		return
	}

	// Vérification format email avec l'utilitaire de validation
	if !utils.ValidateEmail(user.Email) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	// 2. Validation du mot de passe
	if user.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password cannot be empty",
		})
		return
	}

	if len(user.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password must contain at least 6 characters",
		})
		return
	}

	// Vérifier la complexité du mot de passe
	hasLower := strings.ContainsAny(user.Password, "abcdefghijklmnopqrstuvwxyz")
	hasUpper := strings.ContainsAny(user.Password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasDigit := strings.ContainsAny(user.Password, "0123456789")

	if !hasLower || !hasUpper || !hasDigit {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "The password must contain at least one lowercase, one uppercase and one digit",
		})
		return
	}

	// 3. Vérifier si l'email existe déjà
	var existingUser models.User
	if err := db.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		// L'email existe déjà
		c.JSON(http.StatusConflict, gin.H{
			"error": "This email is already used",
		})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Une autre erreur s'est produite lors de la vérification
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error when checking the email existence",
		})
		return
	}

	// 4. Hachage du mot de passe et initialisation des valeurs par défaut
	user.Password = hashPassword(user.Password)
	user.Bio = ""
	user.UserName = ""
	user.Status = models.UserRole
	user.ProfilePicture = ""
	user.StripeCustomerId = ""
	user.SubscriptionPrice = 0
	user.Enable = true
	user.SubscriptionEnable = true
	user.CommentsEnable = true
	user.MessageEnable = true
	user.EmailVerifiedAt = sql.NullTime{Valid: false}
	user.Siret = ""

	// 5. Enregistrement de l'utilisateur dans la base de données
	result := db.DB.Create(&user)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
		"email":   user.Email,
	})
}

func hashPassword(password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPassword)
}
