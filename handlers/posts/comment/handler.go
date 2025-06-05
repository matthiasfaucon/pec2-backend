package comment

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	// Clients connectés au SSE, mappé par postID
	clients = make(map[string]map[chan string]bool)

	// Mutex pour protéger l'accès à la map clients
	clientsMutex sync.RWMutex
)

// Message SSE
type SSEMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// Commentaire à envoyer via SSE
type SSEComment struct {
	ID            string `json:"id"`
	PostID        string `json:"postId"`
	UserID        string `json:"userId"`
	Content       string `json:"content"`
	UserName      string `json:"userName"`
	CreatedAt     string `json:"createdAt"`
	CommentsCount int    `json:"commentsCount"`
}

func GetCommentsByPostID(c *gin.Context) {
	postId := c.Param("id")
	var comments []models.Comment

	if err := db.DB.Where("post_id = ?", postId).Find(&comments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve comments"})
		return
	}

	var commentsResponse []SSEComment
	
	// If no comments, commentsResponse remains an empty slice
	if len(comments) > 0 {
		for _, comment := range comments {
			var user models.User
			db.DB.Select("user_name").Where("id = ?", comment.UserID).First(&user)

			sseComment := SSEComment{
				ID:        comment.ID,
				PostID:    comment.PostID,
				UserID:    comment.UserID,
				Content:   comment.Content,
				UserName:  user.UserName,
				CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}
			commentsResponse = append(commentsResponse, sseComment)
		}
	}

	fmt.Println("Comments retrieved:", commentsResponse)

	c.JSON(http.StatusOK, gin.H{"comments": commentsResponse})
}

// @Summary Handle SSE connection for comments
// @Description Connect to SSE to receive comments in real-time for a specific post
// @Tags comments
// @Param id path string true "Post ID"
// @Param token query string false "JWT Token for web clients (optional)"
// @Security BearerAuth
// @Success 200 {object} map[string]string "Connected to SSE"
// @Failure 400 {object} map[string]string "error: Invalid post ID"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error setting up SSE"
// @Router /posts/{id}/comments/sse [get]
func HandleSSE(c *gin.Context) {
	postID := c.Param("id")

	// Pour l'instant j'ai pas trouver comment passer de header
	// Donc je vais la vérif dans l'URL
	tokenFromQuery := c.Query("token")
	_, exists := c.Get("user_id")

	// Si l'ID utilisateur n'a pas été défini par le middleware (car param dans URL) mais qu'un token est présent dans l'URL
	if !exists && tokenFromQuery != "" {
		_, err := utils.DecodeJWT(tokenFromQuery)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token in URL"})
			return
		}
		exists = true
	}

	// C'est une vérif en plus pour le cas où le token est pas valide
	// dans l'idée c'est toujours à false au début car je passe pas par le middleware
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	var post models.Post

	if err := db.DB.First(&post, "id = ?", postID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		return
	}

	// ça c'est pour les en-têtes pour le SSE
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// là je crée un channel pour chaque client qui se connecte
	messageChan := make(chan string)

	// Là je mets le channel dans la map clients
	// clients est mappé par postID et chaque postID a un channel
	// Et je bloque l'accès à la map clients pendant les modifs
	clientsMutex.Lock()
	if clients[postID] == nil {
		clients[postID] = make(map[chan string]bool)
	}
	clients[postID][messageChan] = true
	clientsMutex.Unlock()

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Streaming not supported"})
		return
	}

	c.Writer.Write([]byte("event: connected\ndata: {\"status\":\"connected\"}\n\n"))

	flusher.Flush()

	var comments []models.Comment
	if err := db.DB.Where("post_id = ?", postID).Find(&comments).Error; err != nil {
		log.Printf("Error retrieving comments: %v", err)
	} else {
		for _, comment := range comments {
			var user models.User
			db.DB.Select("user_name").Where("id = ?", comment.UserID).First(&user)

			sseComment := SSEComment{
				ID:        comment.ID,
				PostID:    comment.PostID,
				UserID:    comment.UserID,
				Content:   comment.Content,
				UserName:  user.UserName,
				CreatedAt: comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			}

			msg := SSEMessage{
				Type:    "existing_comment",
				Payload: sseComment,
			}

			jsonData, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshaling SSE message: %v", err)
				continue
			}

			c.Writer.Write(fmt.Appendf(nil, "event: comment\ndata: %s\n\n", jsonData))
			flusher.Flush()
		}
	}

	// dans la méthode Context c'est écrit
	// "For incoming server requests, the context is canceled when the client's connection closes,..."
	ctx := c.Request.Context() // Se déclenchera lorsque le client se déconnecte

	defer func() {
		clientsMutex.Lock()
		delete(clients[postID], messageChan)
		if len(clients[postID]) == 0 {
			delete(clients, postID)
		}
		clientsMutex.Unlock()
		close(messageChan)
	}()

	for {
		select {
		case message, ok := <-messageChan:
			if !ok {
				return
			}
			c.Writer.Write([]byte(message))
			flusher.Flush()
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
			c.Writer.Write([]byte("event: ping\ndata: {}\n\n"))
			flusher.Flush()
		}
	}
}

// @Summary Create a new comment for a post
// @Description Create a new comment and broadcast it via SSE
// @Tags comments
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param token query string false "JWT Token for web clients (optional)"
// @Param comment body map[string]string true "Comment content"
// @Security BearerAuth
// @Success 201 {object} map[string]string "Comment created"
// @Failure 400 {object} map[string]string "error: Invalid request"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Server error"
// @Router /posts/{id}/comments [post]
func CreateComment(c *gin.Context) {
	postID := c.Param("id")

	// Vérifier si le token est passé en paramètre d'URL pour les clients web
	tokenFromQuery := c.Query("token")
	userID, exists := c.Get("user_id")

	// Si l'ID utilisateur n'a pas été défini par le middleware (car URL) mais qu'un token est présent dans l'URL
	if !exists && tokenFromQuery != "" {
		// Décoder le token de l'URL
		claims, err := utils.DecodeJWT(tokenFromQuery)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token in URL"})
			return
		}
		userID = claims["user_id"]
		exists = true
	}

	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found in token"})
		return
	}

	// Récupérer le contenu du commentaire
	var commentData struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.BindJSON(&commentData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment data"})
		return
	}

	// Créer un nouveau commentaire
	comment := models.Comment{
		PostID:  postID,
		UserID:  userID.(string),
		Content: commentData.Content,
	}

	// Enregistrer dans la base de données
	if err := db.DB.Create(&comment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save comment"})
		return
	}

	// 	// Récupérer le nombre de commentaires pour le post
	var count int64
	if err := db.DB.Model(&models.Comment{}).Where("post_id = ?", postID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count comments"})
		return
	}
	comment.CommentsCount = int(count)

	// Récupérer le nom d'utilisateur
	var user models.User
	db.DB.Select("user_name").Where("id = ?", userID).First(&user)
	// Créer la réponse SSE
	sseComment := SSEComment{
		ID:            comment.ID,
		PostID:        comment.PostID,
		UserID:        comment.UserID,
		Content:       comment.Content,
		UserName:      user.UserName,
		CreatedAt:     comment.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CommentsCount: comment.CommentsCount,
	}
	// Diffuser à tous les clients connectés pour ce post
	broadcastComment(postID, sseComment)

	// Répondre au client
	c.JSON(http.StatusCreated, gin.H{"comment": sseComment})
}

// Diffuser un commentaire à tous les clients connectés pour un post spécifique
func broadcastComment(postID string, comment SSEComment) {
	msg := SSEMessage{
		Type:    "new_comment",
		Payload: comment,
	}

	// Sérialiser le message en JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshaling SSE message: %v", err)
		return
	}

	// Construire le message SSE
	sseData := fmt.Sprintf("event: comment\ndata: %s\n\n", jsonData)

	clientsMutex.RLock()
	defer clientsMutex.RUnlock()

	// Si aucun client n'est connecté pour ce post, on sort
	if _, exists := clients[postID]; !exists {
		return
	}

	// Envoyer le message à tous les clients connectés pour ce post
	for clientChan := range clients[postID] {
		select {
		case clientChan <- sseData:
		default:
			log.Printf("Error broadcasting comment: channel full or closed")
			clientsMutex.Lock()
			delete(clients[postID], clientChan)
			clientsMutex.Unlock()
		}
	}
}
