package email

import (
	"EduSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
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
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// для регистрации userID может быть 0 — возьмём адрес из тела запроса
	userID := c.GetInt("user_id")
	email := ""
	if req.Action == "register" {
		var body struct {
			Email string `json:"email" binding:"required,email"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		email = body.Email
		// можно тут создать черновика user, но опустим
	} else {
		email = c.GetString("email")
	}
	if err := h.svc.RequestCode(c.Request.Context(), userID, email, req.Action); err != nil {
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
	userID := c.GetInt("user_id")
	if err := h.svc.VerifyCode(c.Request.Context(), userID, req.Action, req.Code); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "успешно"})
}
