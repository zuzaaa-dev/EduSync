package middleware

import (
	"EduSync/internal/repository"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func ChatMembershipMiddleware(chatRepo repository.ChatRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		chatIDStr := c.Param("id")
		chatID, err := strconv.Atoi(chatIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор чата"})
			c.Abort()
			return
		}

		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Нет информации о пользователе"})
			c.Abort()
			return
		}
		isTeacher, exists := c.Get("is_teacher")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Нет информации о пользователе"})
			c.Abort()
			return
		}
		var isMember bool
		if isTeacher.(bool) {
			isMember, err = chatRepo.IsOwner(c.Request.Context(), chatID, userID.(int))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки участия"})
				c.Abort()
				return
			}
		} else {
			isMember, err = chatRepo.IsParticipant(c.Request.Context(), chatID, userID.(int))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки участия"})
				c.Abort()
				return
			}
		}

		if !isMember {
			c.JSON(http.StatusForbidden, gin.H{"error": "Вы не являетесь участником данного чата"})
			c.Abort()
			return
		}
		c.Next()
	}
}
