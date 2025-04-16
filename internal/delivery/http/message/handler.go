package chat

import (
	domainChat "EduSync/internal/domain/chat"
	"net/http"
	"strconv"

	serviceChat "EduSync/internal/service"
	"github.com/gin-gonic/gin"
)

type MessageHandler struct {
	messageService serviceChat.MessageService
}

// NewMessageHandler создает новый MessageHandler.
func NewMessageHandler(messageService serviceChat.MessageService) *MessageHandler {
	return &MessageHandler{messageService: messageService}
}

// GetMessagesHandler возвращает список сообщений для чата с поддержкой пагинации.
// Пример запроса: GET /chats/:id/messages?limit=10&offset=0
func (h *MessageHandler) GetMessagesHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	messages, err := h.messageService.GetMessages(c.Request.Context(), chatID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить сообщения"})
		return
	}
	c.JSON(http.StatusOK, messages)
}

// SendMessageHandler создает новое сообщение.
// Пример запроса: POST /chats/:id/messages
func (h *MessageHandler) SendMessageHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	var req struct {
		Text            *string `json:"text"`
		MessageGroupID  *int    `json:"message_group_id"`
		ParentMessageID *int    `json:"parent_message_id"`
		// Файлы на данный момент обрабатываются отдельно
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Нет информации о пользователе"})
		return
	}

	message := domainChat.Message{
		ChatID:          chatID,
		UserID:          userID.(int),
		Text:            req.Text,
		MessageGroupID:  req.MessageGroupID,
		ParentMessageID: req.ParentMessageID,
	}
	messageID, err := h.messageService.SendMessage(c.Request.Context(), message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сообщение"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Сообщение создано", "message_id": messageID})
}

// DeleteMessageHandler удаляет сообщение.
// Пример запроса: DELETE /chats/:id/messages/:messageID
func (h *MessageHandler) DeleteMessageHandler(c *gin.Context) {
	messageIDStr := c.Param("messageID")
	messageID, err := strconv.Atoi(messageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор сообщения"})
		return
	}

	// Предполагаем, что проверка прав осуществляется в сервисном слое.
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Нет информации о пользователе"})
		return
	}
	c.Copy()
	err = h.messageService.DeleteMessage(c.Copy(), messageID, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить сообщение"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Сообщение удалено"})
}

// ReplyMessageHandler создает ответ на сообщение.
// Пример запроса: POST /chats/:id/messages/:messageID/reply
func (h *MessageHandler) ReplyMessageHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}
	parentMessageIDStr := c.Param("messageID")
	parentMessageID, err := strconv.Atoi(parentMessageIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор родительского сообщения"})
		return
	}

	var req struct {
		Text *string `json:"text"`
		// Допустимо добавлять поле для файлов, если необходимо.
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Нет информации о пользователе"})
		return
	}

	message := domainChat.Message{
		ChatID:          chatID, // Можно получить из контекста или параметров URL, если требуется.
		UserID:          userID.(int),
		Text:            req.Text,
		ParentMessageID: &parentMessageID,
	}

	messageID, err := h.messageService.ReplyMessage(c.Request.Context(), parentMessageID, message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать ответ"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Ответ создан", "message_id": messageID})
}

// SearchMessagesHandler ищет сообщения по запросу.
// Пример запроса: GET /chats/:id/messages/search?query=текст&limit=10&offset=0
func (h *MessageHandler) SearchMessagesHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}
	query := c.Query("query")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	messages, err := h.messageService.SearchMessages(c.Request.Context(), chatID, query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка поиска сообщений"})
		return
	}
	c.JSON(http.StatusOK, messages)
}
