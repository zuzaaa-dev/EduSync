package chat

import (
	chatDTO "EduSync/internal/delivery/dto/chat"
	domainChat "EduSync/internal/domain/chat"
	"net/http"
	"strconv"

	"EduSync/internal/service"
	"github.com/gin-gonic/gin"
)

// ChatHandler обрабатывает HTTP-запросы, связанные с чатами.
type ChatHandler struct {
	chatService service.ChatService
}

// NewChatHandler создает новый ChatHandler.
func NewChatHandler(chatService service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// CreateChatHandler создает чат
// @Summary      Создать чат
// @Description  Создает новый чат для группы и предмета (только для преподавателей)
// @Tags         Chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input  body      object{group_id=int,subject_id=int}  true  "Данные чата"
// @Success      201  {object}  object{message=string,chat_id=int}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      401  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /chats [post]
func (h *ChatHandler) CreateChatHandler(c *gin.Context) {
	var req chatDTO.CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	ownerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неизвестный владелец"})
		return
	}

	chat := domainChat.Chat{
		GroupID:   req.GroupID,
		OwnerID:   ownerID.(int),
		SubjectID: req.SubjectID,
	}
	chatInfo, err := h.chatService.CreateChat(c.Request.Context(), chat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать чат"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Чат создан",
		"chat_info": chatDTO.ConvertChatToDTO(chatInfo),
	})
}

// UpdateInviteHandler обновляет ссылку-приглашение
// @Summary      Обновить приглашение
// @Description  Генерирует новую ссылку-приглашение для чата (только владелец)
// @Tags         Chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "ID чата"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      401  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /chats/{id}/invite [put]
func (h *ChatHandler) UpdateInviteHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	ownerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неизвестный владелец"})
		return
	}

	isTeacher := c.GetBool("is_teacher")
	if !isTeacher {
		c.JSON(http.StatusForbidden, gin.H{"error": "Неизвестный владелец"})
		return
	}
	chat, err := h.chatService.RecreateInvite(c.Request.Context(), chatID, ownerID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить приглашение"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Приглашение обновлено",
		"chat":    chat,
	})
}

// JoinChatHandler присоединиться к чату
// @Summary      Присоединиться к чату
// @Description  Позволяет студенту присоединиться к чату по ссылке-приглашению
// @Tags         Chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "ID чата"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      403  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /chats/{id}/join [post]
func (h *ChatHandler) JoinChatHandler(c *gin.Context) {
	isTeacher, exists := c.Keys["is_teacher"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка авторизации"})
		return
	}
	if isTeacher.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Вы не можете присоединиться к группе"})
		return
	}

	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неизвестный участник"})
		return
	}

	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Требуется поле code"})
		return
	}

	chatInfo, err := h.chatService.JoinChat(c.Request.Context(), chatID, userID.(int), body.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Вы успешно присоединились к чату",
		"chat_info": chatInfo})
}

// DeleteChatHandler удаляет чат
// @Summary      Удалить чат
// @Description  Удаляет чат (только для владельца)
// @Tags         Chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "ID чата"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      403  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /chats/{id} [delete]
func (h *ChatHandler) DeleteChatHandler(c *gin.Context) {
	isTeacher, exists := c.Keys["is_teacher"]
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка авторизации"})
		return
	}
	if !isTeacher.(bool) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Вы не можете удалить группу"})
		return
	}

	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	ownerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неизвестный владелец"})
		return
	}

	err = h.chatService.DeleteChat(c.Request.Context(), chatID, ownerID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить чат"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Чат удален"})
}

// GetParticipantsHandler получает участников чата
// @Summary      Участники чата
// @Description  Возвращает список участников чата
// @Tags         Chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "ID чата"
// @Success      200  {array}   Participant
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /chats/{id}/participants [get]
func (h *ChatHandler) GetParticipantsHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	participants, err := h.chatService.ChatParticipants(c.Request.Context(), chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список участников"})
		return
	}
	c.JSON(http.StatusOK, participants)
}

// RemoveParticipantHandler удаляет участника
// @Summary      Удалить участника
// @Description  Удаляет участника из чата (только владелец)
// @Tags         Chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id      path  int  true  "ID чата"
// @Param        userID  path  int  true  "ID участника"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      403  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /chats/{id}/participants/{userID} [delete]
func (h *ChatHandler) RemoveParticipantHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	participantIDStr := c.Param("userID")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}
	participantID, err := strconv.Atoi(participantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор участника"})
		return
	}

	ownerID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неизвестный владелец"})
		return
	}

	err = h.chatService.RemoveParticipant(c.Request.Context(), chatID, ownerID.(int), participantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить участника"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Участник удален"})
}

// LeaveChatHandler покинуть чат
// @Summary      Покинуть чат
// @Description  Позволяет участнику покинуть чат
// @Tags         Chats
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "ID чата"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /chats/{id}/leave [post]
func (h *ChatHandler) LeaveChatHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неизвестный участник"})
		return
	}

	err = h.chatService.LeaveChat(c.Request.Context(), chatID, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось покинуть чат"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Вы покинули чат"})
}

// ListChatsHandler отдаёт все чаты для текущего пользователя.
// GET /chats
func (h *ChatHandler) ListChatsHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	isTeacher := c.GetBool("is_teacher")

	chats, err := h.chatService.ListForUser(c.Request.Context(), userID, isTeacher)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, chats)
}
