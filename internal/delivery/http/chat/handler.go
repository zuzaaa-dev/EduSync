package chat

import (
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

// CreateChatHandler создает новый чат (POST /chats).
func (h *ChatHandler) CreateChatHandler(c *gin.Context) {
	var req struct {
		GroupID   int `json:"group_id" binding:"required"`
		SubjectID int `json:"subject_id" binding:"required"`
	}
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
	chatID, err := h.chatService.CreateChat(c.Request.Context(), chat)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать чат"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Чат создан", "chat_id": chatID})
}

// UpdateInviteHandler позволяет владельцу пересоздать invite-ссылку.
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

	err = h.chatService.RecreateInvite(c.Request.Context(), chatID, ownerID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить приглашение"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Приглашение обновлено"})
}

// JoinChatHandler позволяет пользователю присоединиться к чату (POST /chats/:id/join).
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

	err = h.chatService.JoinChat(c.Request.Context(), chatID, userID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось присоединиться к чату"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Вы успешно присоединились к чату"})
}

// DeleteChatHandler удаляет чат (только для владельца).
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

	chatIDStr := c.Param("chatID")
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

// GetParticipantsHandler возвращает список участников чата (GET /chats/:id/participants).
func (h *ChatHandler) GetParticipantsHandler(c *gin.Context) {
	chatIDStr := c.Param("id")
	chatID, err := strconv.Atoi(chatIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
		return
	}

	participants, err := h.chatService.GetChatParticipants(c.Request.Context(), chatID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить список участников"})
		return
	}
	c.JSON(http.StatusOK, participants)
}

// RemoveParticipantHandler удаляет участника (только для владельца).
func (h *ChatHandler) RemoveParticipantHandler(c *gin.Context) {
	chatIDStr := c.Param("chatID")
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

// LeaveChatHandler – участник покидает чат.
func (h *ChatHandler) LeaveChatHandler(c *gin.Context) {
	chatIDStr := c.Param("chatID")
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
