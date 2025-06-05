package posts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// @Summary Create a new post
// @Description Create a new post with the provided information
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Param name formData string true "Post name"
// @Param isFree formData boolean false "Is the post free"
// @Param enable formData boolean false "Is the post enabled"
// @Param categories formData []string false "Category IDs"
// @Param file formData file false "Post picture"
// @Security BearerAuth
// @Success 201 {object} models.Post
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts [post]
func CreatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in CreatePost")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	name := c.Request.FormValue("name")
	if name == "" {
		utils.LogError(nil, "Name is required in CreatePost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}

	isFreeStr := c.Request.FormValue("isFree")
	var isFree bool
	switch isFreeStr {
	case "true":
		isFree = true
	case "false":
		isFree = false
	default:
		isFree = false

	}

	categoryIDs := c.PostFormArray("categories")
	if len(categoryIDs) == 0 {
		categoriesStr := c.Request.FormValue("categories")
		if categoriesStr != "" {
			if err := json.Unmarshal([]byte(categoriesStr), &categoryIDs); err != nil {
				utils.LogError(err, "Invalid categories format in CreatePost")
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid categories format: " + err.Error()})
				return
			}
		}
	}

	post := models.Post{
		UserID: userID.(string),
		Name:   name,
		IsFree: isFree,
		Enable: true,
	}

	file, err := c.FormFile("file")
	if err == nil && file != nil {
		imageURL, err := utils.UploadImage(file, "post_pictures", "post")
		if err != nil {
			utils.LogError(err, "Error uploading picture in CreatePost")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading picture: " + err.Error()})
			return
		}
		post.PictureURL = imageURL
	} else {
		utils.LogError(nil, "Picture is required in CreatePost")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Picture is required"})
		return
	}

	if len(categoryIDs) > 0 {
		var categories []models.Category
		if err := db.DB.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
			utils.LogError(err, "Error finding categories in CreatePost")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding categories: " + err.Error()})
			return
		}

		if len(categories) == 0 {
			utils.LogError(nil, "No valid categories found in CreatePost")
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid categories found"})
			return
		}

		post.Categories = categories
	}

	if err := db.DB.Create(&post).Error; err != nil {
		utils.LogError(err, "Error creating post in CreatePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating post: " + err.Error()})
		return
	}

	//! C'est à moitié useless, mais c'est pour renvoyer les catégories sinon je les voient pas dans la réponse
	if err := db.DB.Preload("Categories").Where("id = ?", post.ID).First(&post).Error; err != nil {
		utils.LogError(err, "Error retrieving created post in CreatePost")
		fmt.Println("Error retrieving created post:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving created post: " + err.Error()})
		return
	}

	utils.LogSuccessWithUser(userID, "Post created successfully in CreatePost")
	c.JSON(http.StatusCreated, post)
}

// @Summary Get all posts
// @Description Retrieve all posts with optional filtering
// @Tags posts
// @Produce json
// @Param isFree query boolean false "Filter by free posts"
// @Param category query string false "Filter by category ID"
// @Success 200 {array} models.PostResponse
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts [get]
func GetAllPosts(c *gin.Context) {
	var posts []models.Post
	query := db.DB.Preload("Categories").Order("created_at DESC")

	// Filtre pour les posts gratuits/payants
	if isFree := c.Query("isFree"); isFree != "" {
		query = query.Where("is_free = ?", isFree == "true")
	}

	// Afficher le user qui a créé le post
	query = query.Preload("User")

	// Filtre par catégorie
	if categoryID := c.Query("category"); categoryID != "" {
		query = query.Joins("JOIN post_categories ON posts.id = post_categories.post_id").
			Where("post_categories.category_id = ?", categoryID)
	}

	if err := query.Find(&posts).Error; err != nil {
		utils.LogError(err, "Error retrieving posts in GetAllPosts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving posts: " + err.Error()})
		return
	}
	var response []models.PostResponse = make([]models.PostResponse, 0, len(posts))
	for _, post := range posts {
		// Compter le nombre de likes
		var likesCount int64
		db.DB.Model(&models.Like{}).Where("post_id = ?", post.ID).Count(&likesCount)
		fmt.Println("Likes count for post ID", post.ID, ":", likesCount)
		// Compter le nombre de commentaires
		var commentsCount int64
		db.DB.Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&commentsCount)

		// Compter le nombre de reports
		var reportsCount int64
		db.DB.Model(&models.Report{}).Where("post_id = ?", post.ID).Count(&reportsCount)

		// Créer la réponse pour ce post
		postResponse := models.PostResponse{
			ID:         post.ID,
			Name:       post.Name,
			PictureURL: post.PictureURL,
			IsFree:     post.IsFree,
			Enable:     post.Enable,
			Categories: post.Categories,
			CreatedAt:  post.CreatedAt,
			UpdatedAt:  post.UpdatedAt,
			User: models.UserInfo{
				ID:             post.User.ID,
				UserName:       post.User.UserName,
				ProfilePicture: post.User.ProfilePicture,
			},
			LikesCount:    int(likesCount),
			CommentsCount: int(commentsCount),
			ReportsCount:  int(reportsCount),
		}

		response = append(response, postResponse)
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = "0"
	}
	utils.LogSuccessWithUser(userID, "Posts retrieved successfully in GetAllPosts")
	utils.LogSuccess("Posts retrieved successfully in GetAllPosts")
	c.JSON(http.StatusOK, response)
}

// @Summary Get a post by ID
// @Description Retrieve a post by its ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} models.PostResponse
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id} [get]
func GetPostByID(c *gin.Context) {
	var post models.Post
	postID := c.Param("id")

	if err := db.DB.Preload("Categories").Preload("User").First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in GetPostByID")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Compter le nombre de likes
	var likesCount int64
	db.DB.Model(&models.Like{}).Where("post_id = ?", post.ID).Count(&likesCount)

	// Compter le nombre de commentaires
	var commentsCount int64
	db.DB.Model(&models.Comment{}).Where("post_id = ?", post.ID).Count(&commentsCount)

	// Compter le nombre de reports
	var reportsCount int64
	db.DB.Model(&models.Report{}).Where("post_id = ?", post.ID).Count(&reportsCount)

	// Créer la réponse pour ce post
	postResponse := models.PostResponse{
		ID:         post.ID,
		Name:       post.Name,
		PictureURL: post.PictureURL,
		IsFree:     post.IsFree,
		Enable:     post.Enable,
		Categories: post.Categories,
		CreatedAt:  post.CreatedAt,
		UpdatedAt:  post.UpdatedAt,
		User: models.UserInfo{
			ID:             post.User.ID,
			UserName:       post.User.UserName,
			ProfilePicture: post.User.ProfilePicture,
		},
		LikesCount:    int(likesCount),
		CommentsCount: int(commentsCount),
		ReportsCount:  int(reportsCount),
	}

	utils.LogSuccess("Post retrieved successfully in GetPostByID")
	c.JSON(http.StatusOK, postResponse)
}

// @Summary Update a post
// @Description Update a post with the provided information
// @Tags posts
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Post ID"
// @Param name formData string false "Post name"
// @Param isFree formData boolean false "Is the post free"
// @Param enable formData boolean false "Is the post enabled"
// @Param categories formData []string false "Category IDs"
// @Param file formData file false "Post picture"
// @Security BearerAuth
// @Success 200 {object} models.Post
// @Failure 400 {object} map[string]string "error: Invalid input"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id} [put]
func UpdatePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in UpdatePost")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var post models.Post
	postID := c.Param("id")

	if err := db.DB.Preload("Categories").First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in UpdatePost")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Vérifier que l'utilisateur est propriétaire du post ou admin
	userRole, _ := c.Get("user_role")
	if post.UserID != userID.(string) && userRole.(string) != string(models.AdminRole) {
		utils.LogError(nil, "Not authorized to update this post in UpdatePost")
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this post"})
		return
	}

	name := c.Request.FormValue("name")
	isFreeStr := c.Request.FormValue("isFree")
	enableStr := c.Request.FormValue("enable")
	categoriesStr := c.Request.FormValue("categories")

	if name != "" {
		post.Name = name
	}

	if isFreeStr != "" {
		post.IsFree = isFreeStr == "true"
	}

	if enableStr != "" {
		post.Enable = enableStr == "true"
	}

	file, err := c.FormFile("file")
	if err == nil && file != nil {
		if post.PictureURL != "" {
			_ = utils.DeleteImage(post.PictureURL)
		}

		imageURL, err := utils.UploadImage(file, "post_pictures", "post")
		if err != nil {
			utils.LogError(err, "Error uploading picture in UpdatePost")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading picture: " + err.Error()})
			return
		}
		post.PictureURL = imageURL
	}

	if categoriesStr != "" {
		categoryIDs := strings.Split(categoriesStr, ",")
		var categories []models.Category
		if err := db.DB.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
			utils.LogError(err, "Error finding categories in UpdatePost")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding categories: " + err.Error()})
			return
		}

		if len(categories) > 0 {
			if err := db.DB.Model(&post).Association("Categories").Replace(&categories); err != nil {
				utils.LogError(err, "Error updating categories in UpdatePost")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating categories: " + err.Error()})
				return
			}
		}
	}

	if err := db.DB.Save(&post).Error; err != nil {
		utils.LogError(err, "Error updating post in UpdatePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error updating post: " + err.Error()})
		return
	}

	if err := db.DB.Preload("Categories").First(&post, "id = ?", post.ID).Error; err != nil {
		utils.LogError(err, "Error retrieving updated post in UpdatePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving updated post: " + err.Error()})
		return
	}

	utils.LogSuccess("Post updated successfully in UpdatePost")
	c.JSON(http.StatusOK, post)
}

// @Summary Delete a post
// @Description Delete a post by its ID
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Post deleted successfully"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Post not found"
// @Failure 500 {object} map[string]string "error: Error message"
// @Router /posts/{id} [delete]
func DeletePost(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not found in token in DeletePost")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var post models.Post
	postID := c.Param("id")

	if err := db.DB.First(&post, "id = ?", postID).Error; err != nil {
		utils.LogError(err, "Post not found in DeletePost")
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// Vérifier que l'utilisateur est propriétaire du post ou admin
	userRole, _ := c.Get("user_role")
	if post.UserID != userID.(string) && userRole.(string) != string(models.AdminRole) {
		utils.LogError(nil, "Not authorized to delete this post in DeletePost")
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to delete this post"})
		return
	}

	if post.PictureURL != "" {
		_ = utils.DeleteImage(post.PictureURL)
	}

	if err := db.DB.Model(&post).Association("Categories").Clear(); err != nil {
		utils.LogError(err, "Error removing post categories in DeletePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error removing post categories: " + err.Error()})
		return
	}

	if err := db.DB.Delete(&post).Error; err != nil {
		utils.LogError(err, "Error deleting post in DeletePost")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error deleting post: " + err.Error()})
		return
	}

	utils.LogSuccess("Post deleted successfully in DeletePost")
	c.JSON(http.StatusOK, gin.H{"message": "Post deleted successfully"})
}
