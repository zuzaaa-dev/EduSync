package middleware

import (
	"net/http"
	"strings"

	"EduSync/internal/repository/user"
	"EduSync/internal/util"
	"github.com/gin-gonic/gin"
)

// JWTMiddleware проверяет JWT и сверяет его с базой данных.
func JWTMiddleware(tokenRepo *user.TokenRepository, jwtManager *util.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
			c.Abort()
			return
		}

		// Ожидаем формат: "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат токена"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		// Парсим JWT
		claims, err := jwtManager.ParseJWT(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный или просроченный токен"})
			c.Abort()
			return
		}

		// Проверяем в БД, есть ли этот токен (защита от использования удалённых токенов)
		exists, err := tokenRepo.IsTokenValid(tokenStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сервера"})
			c.Abort()
			return
		}
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен отозван"})
			c.Abort()
			return
		}

		// Пробрасываем данные пользователя в контекст
		c.Set("user_id", claims.ID)
		c.Set("is_teacher", claims.IsTeacher)
		c.Set("email", claims.Email)
		c.Set("full_name", claims.FullName)
		c.Next()
	}
}
