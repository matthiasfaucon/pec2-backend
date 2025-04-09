package utils

import (
	"github.com/gin-gonic/gin"
)

// Response structure standard pour les réponses API
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SendSuccess envoie une réponse de succès
func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendError envoie une réponse d'erreur
func SendError(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// ValidateRequestBody vérifie si le body de la requête est valide
func ValidateRequestBody(c *gin.Context, obj interface{}) bool {
	if err := c.ShouldBindJSON(obj); err != nil {
		SendError(c, 400, "Invalid request body: "+err.Error())
		return false
	}
	return true
}
