package email

import (
	"EduSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ConfirmationHandler struct {
	svc service.ConfirmationService
}

func NewConfirmationHandler(svc service.ConfirmationService) *ConfirmationHandler {
	return &ConfirmationHandler{svc: svc}
}

// RequestCode
// @Summary  Запросить код подтверждения
// @Description  Отправляет код на почту для указанного действия
// @Tags         Confirmation
// @Accept       json
// @Produce      json
// @Param        input  body  object{action=string}  true "Действие"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Router       /confirm/request [post]
func (h *ConfirmationHandler) RequestCode(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=register reset_password delete_account"`
		Email  string `json:"email" binding:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RequestCode(c.Request.Context(), req.Email, req.Action); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "код отправлен"})
}

// VerifyCode
// @Summary  Подтвердить код
// @Description  Проверяет код и выполняет действие
// @Tags         Confirmation
// @Accept       json
// @Produce      json
// @Param        input  body  object{action=string,code=string}  true "действие и код"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  object{error=string}
// @Router       /confirm/verify [post]
func (h *ConfirmationHandler) VerifyCode(c *gin.Context) {
	var req struct {
		Action string `json:"action" binding:"required,oneof=register reset_password delete_account"`
		Code   string `json:"code" binding:"required,len=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.VerifyCode(c.Request.Context(), req.Action, req.Code, nil); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "успешно"})
}

// Activate
// @Summary Активировать аккаунт
// @Description По ссылке из письма пользователь активирует свою учётку
// @Tags Confirmation
// @Accept  json
// @Produce  json
// @Param user_id query int true "ID пользователя"
// @Param code    query string true "Код активации"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} dto.ErrorResponse
// @Router /confirm/activate [get]
func (h *ConfirmationHandler) Activate(c *gin.Context) {
	uid, err := strconv.Atoi(c.Query("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}
	if err := h.svc.VerifyCode(c.Request.Context(), "register", code, &uid); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "account activated"})
}

// ResetPassword — POST /api/confirm/reset
// @Summary Сброс пароля
// @Description Сбрасывает пароль по коду из письма
// @Tags Confirmation
// @Accept json
// @Produce json
// @Param input body object{code=string,new_password=string} true "Код и новый пароль"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} dto.ErrorResponse
// @Router /confirm/reset [post]
func (h *ConfirmationHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Code        string `json:"code" binding:"required,len=6"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.ResetPassword(c.Request.Context(), req.Code, req.NewPassword); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "password has been reset"})
}
