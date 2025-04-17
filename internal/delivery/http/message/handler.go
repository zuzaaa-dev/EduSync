package chat

import (
	domainChat "EduSync/internal/domain/chat"
	"mime/multipart"
	"net/http"
	"net/url"
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

// SendMessageHandler создаёт новое сообщение вместе с файлами.
// Запрос должен быть multipart/form-data с полем text, необязательными полями
// message_group_id, parent_message_id и любым количеством файлов в ключе files.
func (h *MessageHandler) SendMessageHandler(c *gin.Context) {
	// 1) Парсим chat_id
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	// 2) Получаем user_id из контекста (JWT‑middleware)
	userIDIface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	userID := userIDIface.(int)

	// 3) Multipart‑form
	text := c.PostForm("text")

	var mgid, pmid *int
	if s := c.PostForm("message_group_id"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			mgid = &v
		}
	}
	if s := c.PostForm("parent_message_id"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			pmid = &v
		}
	}

	// 4) Сбор файлов
	form, err := c.MultipartForm()
	if err != nil && err != http.ErrNotMultipart {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ожидается multipart/form-data"})
		return
	}
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["files"]
	}

	// 5) Собираем доменный объект
	msg := domainChat.Message{
		ChatID: chatID,
		UserID: userID,
		Text: func() *string {
			if text != "" {
				return &text
			}
			return nil
		}(),
		MessageGroupID:  mgid,
		ParentMessageID: pmid,
	}

	// 6) Вызываем сервис с транзакцией и атомарным сохранением
	messageID, err := h.messageService.SendMessageWithFiles(c.Request.Context(), msg, files, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сообщение"})
		return
	}

	// 7) Успех
	c.JSON(http.StatusCreated, gin.H{"message_id": messageID})
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
	//// переиспользуем SendMessageHandler, просто вкладываем parent_message_id
	c.Request.PostForm = url.Values{
		"parent_message_id": {c.Param("messageID")},
		// остальные поля (text, files)… уже в POST body
	}
	h.SendMessageHandler(c)
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
