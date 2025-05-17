package categories

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Create a new category
// @Description Create a new category with the provided information
// @Tags categories
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Category name"
// @Param picture formData file true "Category picture"
// @Security BearerAuth
// @Success 201 {object} models.Category
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /categories [post]
func CreateCategory(c *gin.Context) {
	var categoryCreate models.CategoryCreate
	categoryCreate.Name = c.PostForm("name")

	if categoryCreate.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name is required",
		})
		return
	}

	var existingCategory models.Category
	resultInCategories := db.DB.First(&existingCategory, "name = ?", categoryCreate.Name)
	if resultInCategories.Error == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category already exists",
		})
		return
	}

	file, err := c.FormFile("picture")
	if err == nil && file != nil {
		imageURL, err := utils.UploadImage(file, "category_pictures", "category")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading picture: " + err.Error()})
			return
		}
		categoryCreate.PictureURL = imageURL
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Picture is required"})
		return
	}

	category := models.Category{
		Name:       categoryCreate.Name,
		PictureURL: categoryCreate.PictureURL,
	}

	result := db.DB.Create(&category)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating category: " + result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, category)
}

// @Summary Get all categories
// @Description Retrieve all categories
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Category
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /categories [get]
func GetAllCategories(c *gin.Context) {
	var categories []models.Category

	// Create a new database session to avoid prepared statement conflicts
	session := db.DB.Session(&gorm.Session{})
	result := session.Order("name ASC").Find(&categories)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// @Summary Delete a category
// @Description Delete a category by its ID
// @Tags categories
// @Produce json
// @Param id path string true "Category ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Category deleted successfully"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Category not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /categories/{id} [delete]
func DeleteCategory(c *gin.Context) {
	categoryID := c.Param("id")

	// On vérifie que la catégorie existe avant de la supprimer
	var category models.Category
	result := db.DB.First(&category, "id = ?", categoryID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// First, delete all associations in the post_categories table
	if err := db.DB.Exec("DELETE FROM post_categories WHERE category_id = ?", categoryID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing category from posts: " + err.Error()})
		return
	}

	if category.PictureURL != "" {
		_ = utils.DeleteImage(category.PictureURL)
	}

	result = db.DB.Delete(&category)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting category: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
}

// @Summary Update a category
// @Description Update a category with the provided information
// @Tags categories
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Category ID"
// @Param name formData string true "Category name"
// @Param picture formData file true "Category picture"
// @Security BearerAuth
// @Success 200 {object} models.Category
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Category not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /categories/{id} [put]
func UpdateCategory(c *gin.Context) {
	categoryID := c.Param("id")

	var category models.Category
	result := db.DB.First(&category, "id = ?", categoryID)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	// Récupérer le nom depuis le form-data
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	// Vérifier si le nom existe déjà pour une autre catégorie
	var existingCategory models.Category
	if err := db.DB.Where("name = ? AND id != ?", name, categoryID).First(&existingCategory).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category name already exists"})
		return
	}

	category.Name = name

	// Gérer l'upload de l'image
	file, err := c.FormFile("picture")
	if err == nil && file != nil {
		// Supprimer l'ancienne image si elle existe
		if category.PictureURL != "" {
			_ = utils.DeleteImage(category.PictureURL)
		}

		// Upload de la nouvelle image
		imageURL, err := utils.UploadImage(file, "category_pictures", "category")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading picture: " + err.Error()})
			return
		}
		category.PictureURL = imageURL
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Picture is required"})
		return
	}

	result = db.DB.Save(&category)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating category: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}
