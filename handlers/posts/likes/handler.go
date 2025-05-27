package likes

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
)

// @Summary Toggle like on a post
// @Description Add or remove a like on a post
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Like added/removed successfully"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id}/like [post]
func ToggleLike(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in ToggleLike")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	postID := c.Param("id")

	// Vérifier si le post existe
	var post models.Post
	if err := db.DB.First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in ToggleLike")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	var like models.Like
	result := db.DB.Where("post_id = ? AND user_id = ?", postID, userID).First(&like)

	if result.Error == nil {
		// Le like existe déjà, on le supprime
		if err := db.DB.Delete(&like).Error; err != nil {
			utils.LogError(err, "Error removing like in ToggleLike")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing like: " + err.Error()})
			return
		}

		// Récupérer le nombre actuel de likes pour ce post
		var likesCount int64
		db.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&likesCount)

		utils.LogSuccessWithUser(userID, "Like removed successfully in ToggleLike")
		c.JSON(http.StatusOK, gin.H{
			"message":    "Like removed successfully",
			"action":     "removed",
			"likesCount": likesCount,
		})
		return
	}

	// Le like n'existe pas, on le crée
	like = models.Like{
		PostID: postID,
		UserID: userID.(string),
	}

	if err := db.DB.Create(&like).Error; err != nil {
		utils.LogError(err, "Error adding like in ToggleLike")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding like: " + err.Error()})
		return
	}

	// Récupérer le nombre actuel de likes pour ce post
	var likesCount int64
	db.DB.Model(&models.Like{}).Where("post_id = ?", postID).Count(&likesCount)

	utils.LogSuccessWithUser(userID, "Like added successfully in ToggleLike")
	c.JSON(http.StatusOK, gin.H{
		"message":    "Like added successfully",
		"action":     "added",
		"likesCount": likesCount,
	})
}
