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
// @Summary      Регистрация нового пользователя
// @Description  Создаёт аккаунт студента или преподавателя
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input  body      RegistrationUserReq  true  "Данные для регистрации"
// @Success      201    {object}  object{message=string,user_id=int}
// @Failure      400    {object}  ErrorResponse
// @Failure      409    {object}  ErrorResponse
// @Router       /register [post]
func (h *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegistrationUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}
	if req.InstitutionID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Поле institution_id не может быть отрицательным"})
		return
	}
	if req.GroupID <= 0 && !*req.IsTeacher {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Поле group_id не может быть отрицательным"})
		return
	}
	converter, err := req.ConvertToSvc()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка создания пользователя"})
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
// @Summary      Аутентификация пользователя
// @Description  Вход в систему с использованием email и пароля
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input  body      LoginUserReq  true  "Данные для входа"
// @Success      200    {object}  PairTokenResp
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Router       /login [post]
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
		AccessToken:  "Bearer " + accessToken,
		RefreshToken: "Bearer " + refreshToken,
	})
}

// LogoutHandler отзывает токен.
// @Summary      Выход из системы
// @Description  Отзывает текущий токен доступа
// @Tags         Auth
// @Security     BearerAuth
// @Success      200  {object}  object{message=string}
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /logout [post]
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
// @Summary      Обновление токена доступа
// @Description  Обновляет access-токен с использованием refresh-токена
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        input  body      RefreshTokenReq  true  "Refresh токен"
// @Success      200    {object}  PairTokenResp
// @Failure      400    {object}  ErrorResponse
// @Failure      401    {object}  ErrorResponse
// @Router       /refresh [post]
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
		AccessToken:  "Bearer " + accessToken,
		RefreshToken: "Bearer " + newRefreshToken,
	})
}

// ProfileHandler обрабатывает получения профиля.
// @Summary      Получение профиля
// @Description  Получение данных профиля
// @Tags         Auth
// @Security     BearerAuth
// @Success      200  {object}  ProfileResp
// @Failure      401  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /profile [get]
func (h *AuthHandler) ProfileHandler(c *gin.Context) {
	userID, _ := c.Get("user_id")
	email, _ := c.Get("email")
	fullName, _ := c.Get("full_name")
	isTeacher, _ := c.Get("is_teacher")
	groupID, _ := c.Get("group_id")
	institutionID, _ := c.Get("institution_id")
	c.JSON(http.StatusOK, ProfileResp{
		UserID:        userID.(int),
		Email:         email.(string),
		FullName:      fullName.(string),
		InstitutionID: institutionID.(int),
		GroupID:       groupID.(int),
		IsTeacher:     isTeacher.(bool),
	})
}
