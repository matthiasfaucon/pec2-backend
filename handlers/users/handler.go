package users

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary Get all users
// @Description Retrieves a list of all users
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "users: array of user objects"
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

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// @Summary Get user by ID
// @Description Retrieves a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{} "user: user object"
// @Failure 400 {object} map[string]string "error: Invalid user ID"
// @Failure 404 {object} map[string]string "error: User not found"
// @Failure 500 {object} map[string]string "error: error message"
// @Router /users/{id} [get]
func GetUserByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID d'utilisateur invalide"})
		return
	}

	var user models.User
	result := db.DB.First(&user, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Utilisateur non trouvé"})
		return
	}

	// Ne pas envoyer le mot de passe
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{"user": user})
}
