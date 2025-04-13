package users

import (
	"database/sql"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body models.UserCreate true "User information"
// @Success 200 {object} map[string]interface{} "message: User created successfully, email: user email"
// @Failure 400 {object} map[string]interface{} "error: Invalid input"
// @Failure 500 {object} map[string]interface{} "error: Error message"
// @Router /user [post]
func CreateUser(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input",
		})
		return
	}

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
