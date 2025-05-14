package privateMessages

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// @Summary Create a private message
// @Description Send a private message from the authenticated user to another user
// @Tags private-messages
// @Accept json
// @Produce json
// @Param message body models.PrivateMessageCreate true "Message information"
// @Security BearerAuth
// @Success 201 {object} models.PrivateMessage "Created message"
// @Failure 400 {object} map[string]string "error: Invalid request data"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 404 {object} map[string]string "error: Receiver not found"
// @Failure 500 {object} map[string]string "error: Error creating message"
// @Router /private-messages [post]
func CreatePrivateMessage(c *gin.Context) {
	senderID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messageCreate models.PrivateMessageCreate
	if err := c.ShouldBindJSON(&messageCreate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	var receiver models.User
	if result := db.DB.Where("id = ?", messageCreate.ReceiverID).First(&receiver); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying receiver: " + result.Error.Error()})
		}
		return
	}

	if !receiver.MessageEnable {
		c.JSON(http.StatusForbidden, gin.H{"error": "Receiver has disabled private messages"})
		return
	}

	privateMessage := models.PrivateMessage{
		SenderID:   senderID.(string),
		ReceiverID: messageCreate.ReceiverID,
		Content:    messageCreate.Content,
		Status:     "UNREAD",
	}

	if result := db.DB.Create(&privateMessage); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating message: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, privateMessage)
}

// @Summary Get user messages
// @Description Get all messages sent and received by the authenticated user
// @Tags private-messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.PrivateMessage "List of messages"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error retrieving messages"
// @Router /private-messages [get]
func GetUserMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messages []models.PrivateMessage

	result := db.DB.Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&messages)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving messages: " + result.Error.Error()})
		return
	}

	type EnhancedMessage struct {
		models.PrivateMessage
		SenderName    string `json:"senderName"`
		ReceiverName  string `json:"receiverName"`
		IsCurrentUser bool   `json:"isCurrentUser"`
	}

	var enhancedMessages []EnhancedMessage

	for _, msg := range messages {
		var sender, receiver models.User

		db.DB.Select("user_name").Where("id = ?", msg.SenderID).First(&sender)

		db.DB.Select("user_name").Where("id = ?", msg.ReceiverID).First(&receiver)

		enhancedMsg := EnhancedMessage{
			PrivateMessage: msg,
			SenderName:     sender.UserName,
			ReceiverName:   receiver.UserName,
			IsCurrentUser:  msg.SenderID == userID.(string),
		}

		enhancedMessages = append(enhancedMessages, enhancedMsg)
	}

	c.JSON(http.StatusOK, enhancedMessages)
}

// @Summary Get received messages
// @Description Get all messages received by the authenticated user
// @Tags private-messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} object "List of received messages"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error retrieving messages"
// @Router /private-messages/received [get]
func GetReceivedMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messages []models.PrivateMessage

	result := db.DB.Where("receiver_id = ?", userID).
		Order("created_at DESC").
		Find(&messages)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving messages: " + result.Error.Error()})
		return
	}

	type EnhancedMessage struct {
		models.PrivateMessage
		SenderName string `json:"senderName"`
	}

	var enhancedMessages []EnhancedMessage

	for _, msg := range messages {
		var sender models.User

		db.DB.Select("user_name").Where("id = ?", msg.SenderID).First(&sender)

		enhancedMsg := EnhancedMessage{
			PrivateMessage: msg,
			SenderName:     sender.UserName,
		}

		enhancedMessages = append(enhancedMessages, enhancedMsg)
	}

	c.JSON(http.StatusOK, enhancedMessages)
}

// @Summary Get sent messages
// @Description Get all messages sent by the authenticated user
// @Tags private-messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} object "List of sent messages"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 500 {object} map[string]string "error: Error retrieving messages"
// @Router /private-messages/sent [get]
func GetSentMessages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messages []models.PrivateMessage

	result := db.DB.Where("sender_id = ?", userID).
		Order("created_at DESC").
		Find(&messages)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving messages: " + result.Error.Error()})
		return
	}

	type EnhancedMessage struct {
		models.PrivateMessage
		ReceiverName string `json:"receiverName"`
	}

	var enhancedMessages []EnhancedMessage

	for _, msg := range messages {
		var receiver models.User

		db.DB.Select("user_name").Where("id = ?", msg.ReceiverID).First(&receiver)

		enhancedMsg := EnhancedMessage{
			PrivateMessage: msg,
			ReceiverName:   receiver.UserName,
		}

		enhancedMessages = append(enhancedMessages, enhancedMsg)
	}

	c.JSON(http.StatusOK, enhancedMessages)
}
