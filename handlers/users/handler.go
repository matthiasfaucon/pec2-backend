package users

import (
	"database/sql"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// CreateUser save a user in db
// @Summary Users CreateUser
// @Description Endpoint for save a user in db
// @Tags createUser
// @Produce json
// @Success 201 {object} utils.Response
// @Router /users [post]
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

func hashPassword  (password string) string {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashedPassword)
}
