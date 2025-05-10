package chat

import (
	domainChat "EduSync/internal/domain/chat"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

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

// GetMessagesHandler возвращает список сообщений
// @Summary      Получить сообщения
// @Description  Возвращает список сообщений чата с пагинацией
// @Tags         Messages
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path  int  true   "ID чата"
// @Param        limit   query int  false "Лимит сообщений"  default(10)
// @Param        offset  query int  false "Смещение"         default(0)
// @Success      200  {array}   Message
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chats/{id}/messages [get]
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

	messages, err := h.messageService.Messages(c.Request.Context(), chatID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить сообщения"})
		return
	}
	c.JSON(http.StatusOK, messages)
}

// SendMessageHandler отправляет сообщение
// @Summary      Отправить сообщение
// @Description  Создает новое сообщение с возможностью прикрепления файлов
// @Tags         Messages
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id                 path  int     true  "ID чата"
// @Param        text               formData  string  false "Текст сообщения"
// @Param        message_group_id   formData  int     false "ID группы сообщений"
// @Param        parent_message_id  formData  int     false "ID родительского сообщения"
// @Param        files              formData  []file  false "Прикрепленные файлы"
// @Success      201  {object}  object{message_id=int}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chats/{id}/messages [post]
func (h *MessageHandler) SendMessageHandler(c *gin.Context) {
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	userIDIface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	userID := userIDIface.(int)

	text := c.PostForm("text")

	if text != "" && len(text) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "текст не более 1000 символов"})
		return
	}

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

	form, err := c.MultipartForm()
	if err != nil && err != http.ErrNotMultipart {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ожидается multipart/form-data"})
		return
	}
	var files []*multipart.FileHeader
	if form != nil {
		files = form.File["files"]
	}
	if len(files) > 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "максимум 8 файлов на сообщение"})
		return
	}
	const maxFileSize = 30 << 20
	allowedExt := map[string]bool{
		// доки
		".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
		".ppt": true, ".pptx": true, ".pdf": true, ".txt": true,
		// картинки
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	}
	for _, fh := range files {
		if fh.Size > maxFileSize {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("файл %s превышает 30МБ", fh.Filename)})
			return
		}
		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !allowedExt[ext] {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("недопустимый формат файла %s", fh.Filename)})
			return
		}
	}

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

	messageID, err := h.messageService.SendMessageWithFiles(c.Request.Context(), msg, files, c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сообщение"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message_id": messageID})
}

// DeleteMessageHandler удаляет сообщение
// @Summary      Удалить сообщение
// @Description  Удаляет сообщение (доступно только автору)
// @Tags         Messages
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id        path  int  true  "ID чата"
// @Param        messageID path  int  true  "ID сообщения"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chats/{id}/messages/{messageID} [delete]
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

// ReplyMessageHandler отвечает на сообщение
// @Summary      Ответить на сообщение
// @Description  Создает ответ на указанное сообщение
// @Tags         Messages
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        id        path  int  true  "ID чата"
// @Param        messageID path  int  true  "ID сообщения"
// @Param        text      formData  string  false "Текст сообщения"
// @Param        files     formData  []file  false "Прикрепленные файлы"
// @Success      201  {object}  object{message_id=int}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      401  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chats/{id}/messages/{messageID}/reply [post]
func (h *MessageHandler) ReplyMessageHandler(c *gin.Context) {
	c.Request.PostForm = url.Values{
		"parent_message_id": {c.Param("messageID")},
	}
	h.SendMessageHandler(c)
}

// SearchMessagesHandler ищет сообщения
// @Summary      Поиск сообщений
// @Description  Осуществляет полнотекстовый поиск по сообщениям чата
// @Tags         Messages
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path  int     true  "ID чата"
// @Param        query query string  true  "Поисковый запрос"
// @Param        limit   query int     false "Лимит результатов"  default(10)
// @Param        offset  query int     false "Смещение"           default(0)
// @Success      200  {array}   Message
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chats/{id}/messages/search [get]
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

func (h *MessageHandler) UpdateMessageHandler(c *gin.Context) {
	// 1) Параметры
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil || chatID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id чата"})
		return
	}
	messageID, err := strconv.Atoi(c.Param("messageID"))
	if err != nil || messageID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный id сообщения"})
		return
	}
	// 2) Тело запроса
	var req struct {
		Text *string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректное содержание запроса"})
		return
	}
	// 3) Получить user_id
	userID := c.GetInt("user_id")
	// 4) Вызвать сервис
	updated, err := h.messageService.UpdateMessage(c.Copy(), messageID, userID, req.Text)
	if err != nil {
		switch err {
		case domainChat.ErrNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "сообщение не найдено"})
		case domainChat.ErrPermissionDenied:
			c.JSON(http.StatusForbidden, gin.H{"error": "нет прав для редактирования"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	// 5) Ответ
	c.JSON(http.StatusOK, updated)
}
