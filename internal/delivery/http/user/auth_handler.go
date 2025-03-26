package user

import (
	"EduSync/internal/service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AuthHandler обрабатывает запросы, связанные с аутентификацией.
type AuthHandler struct {
	authService service.UserService
}

// NewAuthHandler создает новый обработчик аутентификации.
func NewAuthHandler(authService service.UserService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// RegisterHandler обрабатывает регистрацию.
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegistrationUserReq

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}
	converter, err := req.ConvertToSvc()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, err := h.authService.Register(c.Request.Context(), converter)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Пользователь создан", "user_id": userID})
}

// LoginHandler обрабатывает логин.
func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Для примера берем userAgent и ip из запроса
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	accessToken, refreshToken, err := h.authService.Login(c.Request.Context(), req.Email, req.Password, userAgent, ipAddress)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, PairTokenResp{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// LogoutHandler отзывающий токен.
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен отсутствует"})
		return
	}
	token = strings.TrimPrefix(token, "Bearer ")
	if err := h.authService.Logout(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка выхода"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен"})
}

// RefreshTokenHandler обрабатывает обновление access-токена.
func (h *AuthHandler) RefreshTokenHandler(c *gin.Context) {
	var req RefreshTokenReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	// Извлекаем userAgent и ip из запроса, как в LoginHandler
	userAgent := c.Request.UserAgent()
	ipAddress := c.ClientIP()

	accessToken, newRefreshToken, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken, userAgent, ipAddress)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, PairTokenResp{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	})
}
