package user

import (
	"net/http"
	"strings"

	"EduSync/internal/service/user"
	"github.com/gin-gonic/gin"
)

// AuthHandler обрабатывает запросы, связанные с аутентификацией.
type AuthHandler struct {
	authService *user.AuthService
}

// NewAuthHandler создает новый обработчик аутентификации.
func NewAuthHandler(authService *user.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterHandler обрабатывает регистрацию.
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Password  string `json:"password" binding:"required"`
		FullName  string `json:"full_name" binding:"required"`
		IsTeacher bool   `json:"is_teacher" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	userID, err := h.authService.Register(req.Email, req.Password, req.FullName, req.IsTeacher)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Пользователь создан", "user_id": userID})
}

// LoginHandler обрабатывает логин.
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Для примера берем userAgent и ip из запроса
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	accessToken, refreshToken, err := h.authService.Login(req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken, "refresh_token": refreshToken})
}

// LogoutHandler отзывающий токен.
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен отсутствует"})
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")
	if err := h.authService.Logout(token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка выхода"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен"})
}

// RefreshTokenHandler обрабатывает обновление access-токена.
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Извлекаем userAgent и ip из запроса, как в LoginHandler
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	accessToken, newRefreshToken, err := h.authService.RefreshToken(req.RefreshToken, userAgent, ipAddress)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	})
}
