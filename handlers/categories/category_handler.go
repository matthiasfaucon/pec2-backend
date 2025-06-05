package categories

import (
	"errors"
	"fmt"
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
// @Accept json
// @Produce json
// @Param name formData string true "Category name"
// @Param picture formData file true "Category picture"
// @Param body body models.CategoryCreate false "Category data (for JSON requests)"
// @Security BearerAuth
// @Success 201 {object} models.Category
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /categories [post]
func CreateCategory(c *gin.Context) {
	fmt.Println("CreateCategory called")
	var categoryCreate models.CategoryCreate
	fmt.Println("categoryCreate initialized")

	contentType := c.GetHeader("Content-Type")
	fmt.Println("contentType:", contentType)

	// Handle both JSON and form-data
	if contentType == "application/json" {
		fmt.Println("Handling JSON input")
		if err := c.ShouldBindJSON(&categoryCreate); err != nil {
			utils.LogError(err, "Erreur de binding JSON dans CreateCategory")
			fmt.Println("Error binding JSON:", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid JSON input: " + err.Error(),
			})
			return
		}
		fmt.Println("JSON binding successful:", categoryCreate)
	} else {
		fmt.Println("Handling form data")
		categoryCreate.Name = c.PostForm("name")

		file, err := c.FormFile("picture")
		fmt.Println("File:", file != nil, "Error:", err)
		if err == nil && file != nil {
			imageURL, err := utils.UploadImage(file, "category_pictures", "category")
			if err != nil {
				utils.LogError(err, "Error when uploading picture in CreateCategory")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading picture: " + err.Error()})
				return
			}
			categoryCreate.PictureURL = imageURL
		} else {
			utils.LogError(errors.New("picture manquante"), "Picture is required for form data dans CreateCategory")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Picture is required for form data"})
			return
		}
	}

	if categoryCreate.Name == "" {
		utils.LogError(errors.New("nom manquant"), "Name is required dans CreateCategory")
		fmt.Println("Name is required")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name is required",
		})
		return
	}

	var existingCategory models.Category
	resultInCategories := db.DB.First(&existingCategory, "name = ?", categoryCreate.Name)
	if resultInCategories.Error == nil {
		utils.LogError(errors.New("catégorie déjà existante"), "Category already exists dans CreateCategory")
		fmt.Println("Category already exists")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Category already exists",
		})
		return
	}

	category := models.Category{
		Name:       categoryCreate.Name,
		PictureURL: categoryCreate.PictureURL,
	}
	fmt.Println("Creating category:", category)

	result := db.DB.Create(&category)
	if result.Error != nil {
		utils.LogError(result.Error, "Error when creating category in CreateCategory")
		fmt.Println("Error creating category:", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error creating category: " + result.Error.Error(),
		})
		return
	}

	fmt.Println("Category created successfully:", category)
	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Category created successfully in CreateCategory")
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

	session := db.DB.Session(&gorm.Session{})
	result := session.Order("name ASC").Find(&categories)
	if result.Error != nil {
		utils.LogError(result.Error, "Error when retrieving categories in GetAllCategories")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "List of categories retrieved successfully in GetAllCategories")
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

	var category models.Category
	result := db.DB.First(&category, "id = ?", categoryID)
	if result.Error != nil {
		utils.LogError(result.Error, "Category not found in DeleteCategory")
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := db.DB.Exec("DELETE FROM post_categories WHERE category_id = ?", categoryID).Error; err != nil {
		utils.LogError(err, "Error when removing category from posts in DeleteCategory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing category from posts: " + err.Error()})
		return
	}

	if category.PictureURL != "" {
		_ = utils.DeleteImage(category.PictureURL)
	}

	result = db.DB.Delete(&category)
	if result.Error != nil {
		utils.LogError(result.Error, "Error when deleting category in DeleteCategory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting category: " + result.Error.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Category deleted successfully in DeleteCategory")
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
		utils.LogError(result.Error, "Category not found in UpdateCategory")
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	name := c.PostForm("name")
	if name == "" {
		utils.LogError(errors.New("nom manquant"), "Name is required in UpdateCategory")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	var existingCategory models.Category
	if err := db.DB.Where("name = ? AND id != ?", name, categoryID).First(&existingCategory).Error; err == nil {
		utils.LogError(errors.New("nom de catégorie déjà existant"), "Category name already exists dans UpdateCategory")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category name already exists"})
		return
	}

	category.Name = name

	file, err := c.FormFile("picture")
	if err == nil && file != nil {
		if category.PictureURL != "" {
			_ = utils.DeleteImage(category.PictureURL)
		}

		imageURL, err := utils.UploadImage(file, "category_pictures", "category")
		if err != nil {
			utils.LogError(err, "Error when uploading picture in UpdateCategory")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading picture: " + err.Error()})
			return
		}
		category.PictureURL = imageURL
	} else {
		utils.LogError(errors.New("picture manquante"), "Picture is required dans UpdateCategory")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Picture is required"})
		return
	}

	result = db.DB.Save(&category)
	if result.Error != nil {
		utils.LogError(result.Error, "Error when updating category in UpdateCategory")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating category: " + result.Error.Error()})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Category updated successfully in UpdateCategory")
	c.JSON(http.StatusOK, category)
}
