package privateMessages

import (
	"net/http"
	"pec2-backend/db"
	"pec2-backend/models"
	"pec2-backend/utils"

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
		utils.LogError(nil, "User not authenticated dans CreatePrivateMessage")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messageCreate models.PrivateMessageCreate
	if err := c.ShouldBindJSON(&messageCreate); err != nil {
		utils.LogError(err, "Error binding JSON in CreatePrivateMessage")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data: " + err.Error()})
		return
	}

	var receiver models.User
	if result := db.DB.Where("user_name = ?", messageCreate.ReceiverUserName).First(&receiver); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.LogError(result.Error, "Receiver not found in CreatePrivateMessage")
			c.JSON(http.StatusNotFound, gin.H{"error": "Receiver not found"})
		} else {
			utils.LogError(result.Error, "Error verifying receiver in CreatePrivateMessage")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error verifying receiver: " + result.Error.Error()})
		}
		return
	}

	if !receiver.MessageEnable {
		utils.LogError(nil, "Receiver has disabled private messages in CreatePrivateMessage")
		c.JSON(http.StatusForbidden, gin.H{"error": "Receiver has disabled private messages"})
		return
	}

	privateMessage := models.PrivateMessage{
		SenderID:   senderID.(string),
		ReceiverID: receiver.ID,
		Content:    messageCreate.Content,
		Status:     models.MessageStatusUnread,
	}

	if result := db.DB.Create(&privateMessage); result.Error != nil {
		utils.LogError(result.Error, "Error creating private message in CreatePrivateMessage")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating message: " + result.Error.Error()})
		return
	}

	utils.LogSuccess("Private message created successfully in CreatePrivateMessage")
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
		utils.LogError(nil, "User not authenticated in GetUserMessages")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messages []models.PrivateMessage

	result := db.DB.Where("sender_id = ? OR receiver_id = ?", userID, userID).
		Order("created_at DESC").
		Find(&messages)

	if result.Error != nil {
		utils.LogError(result.Error, "Error retrieving messages in GetUserMessages")
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

	utils.LogSuccess("User messages retrieved successfully in GetUserMessages")
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
		utils.LogError(nil, "User not authenticated in GetReceivedMessages")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messages []models.PrivateMessage

	result := db.DB.Where("receiver_id = ?", userID).
		Order("created_at DESC").
		Find(&messages)

	if result.Error != nil {
		utils.LogError(result.Error, "Error retrieving received messages in GetReceivedMessages")
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

	utils.LogSuccess("Received messages retrieved successfully in GetReceivedMessages")
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
		utils.LogError(nil, "User not authenticated in GetSentMessages")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var messages []models.PrivateMessage

	result := db.DB.Where("sender_id = ?", userID).
		Order("created_at DESC").
		Find(&messages)

	if result.Error != nil {
		utils.LogError(result.Error, "Error retrieving sent messages in GetSentMessages")
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

	utils.LogSuccess("Sent messages retrieved successfully in GetSentMessages")
	c.JSON(http.StatusOK, enhancedMessages)
}

// @Summary Mark message as read
// @Description Mark a specific private message as read
// @Tags private-messages
// @Accept json
// @Produce json
// @Param id path string true "Message ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string "message: Message marked as read"
// @Failure 400 {object} map[string]string "error: Bad request"
// @Failure 401 {object} map[string]string "error: Unauthorized"
// @Failure 403 {object} map[string]string "error: Forbidden"
// @Failure 404 {object} map[string]string "error: Message not found"
// @Failure 500 {object} map[string]string "error: Error updating message"
// @Router /private-messages/{id}/read [patch]
func MarkMessageAsRead(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		utils.LogError(nil, "User not authenticated in MarkMessageAsRead")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	messageID := c.Param("id")
	if messageID == "" {
		utils.LogError(nil, "Message ID is required in MarkMessageAsRead")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message ID is required"})
		return
	}

	var message models.PrivateMessage
	if result := db.DB.Where("id = ?", messageID).First(&message); result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			utils.LogError(result.Error, "Message not found in MarkMessageAsRead")
			c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		} else {
			utils.LogError(result.Error, "Error retrieving message in MarkMessageAsRead")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving message: " + result.Error.Error()})
		}
		return
	}

	if message.ReceiverID != userID.(string) {
		utils.LogError(nil, "Permission denied to mark message as read in MarkMessageAsRead")
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to mark this message as read"})
		return
	}

	if message.Status == models.MessageStatusRead {
		utils.LogSuccess("Message already marked as read in MarkMessageAsRead")
		c.JSON(http.StatusOK, gin.H{"message": "Message is already marked as read"})
		return
	}

	if result := db.DB.Model(&message).Update("status", models.MessageStatusRead); result.Error != nil {
		utils.LogError(result.Error, "Error updating message status in MarkMessageAsRead")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error marking message as read: " + result.Error.Error()})
		return
	}

	utils.LogSuccess("Message marked as read successfully in MarkMessageAsRead")
	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}
