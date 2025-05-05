package categories

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new category
// @Description Create a new category with the provided information
// @Tags categories
// @Accept json
// @Produce json
// @Param category body models.CategoryCreate true "Category information"
// @Security BearerAuth
// @Success 201 {object} models.Category
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /categories [post]
func CreateCategory(c *gin.Context) {
	var categoryCreate models.CategoryCreate
	isWellFormatted := c.ShouldBindJSON(&categoryCreate)
	if isWellFormatted != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid input: " + isWellFormatted.Error(),
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

	category := models.Category{
		Name: categoryCreate.Name,
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
	
	result := db.DB.Order("name ASC").Find(&categories)
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
// @Accept json
// @Produce json
// @Param id path string true "Category ID"
// @Param category body models.CategoryCreate true "Updated category information"
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

	var categoryUpdate models.CategoryCreate
	isWellFormatted := c.ShouldBindJSON(&categoryUpdate)
	if isWellFormatted != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + isWellFormatted.Error()})
		return
	}

	category.Name = categoryUpdate.Name

	result = db.DB.Save(&category)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating category: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, category)
}