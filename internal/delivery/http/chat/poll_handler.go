package chat

import (
	"EduSync/internal/delivery/http/chat/dto"
	"net/http"
	"strconv"

	srv "EduSync/internal/service"
	"github.com/gin-gonic/gin"
)

type PollHandler struct {
	svc srv.PollService
}

func NewPollHandler(srv srv.PollService) *PollHandler {
	return &PollHandler{svc: srv}
}

// CreatePoll
// @Summary      Создать опрос
// @Description  Только владелец чата может.
// @Tags         Polls
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        chat_id  path  int            true  "ID чата"
// @Param        input    body  dto.CreatePollReq  true "Тело запроса"
// @Success      201  {object}  object{poll_id=int}
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /chats/{chat_id}/polls [post]
func (h *PollHandler) CreatePoll(c *gin.Context) {
	chatID, _ := strconv.Atoi(c.Param("id"))
	userID := c.GetInt("user_id")

	var req dto.CreatePollReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	pid, err := h.svc.CreatePoll(c.Request.Context(), userID, chatID, req.Question, req.Options)
	if err != nil {
		code := http.StatusInternalServerError
		if err.Error() == "permission denied" {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"poll_id": pid})
}

// DeletePoll
// @Summary      Удалить опрос
// @Tags         Polls
// @Security     BearerAuth
// @Param        poll_id  path  int  true  "ID опроса"
// @Success      200  {object}  object{message=string}
// @Failure      403  {object}  dto.ErrorResponse
// @Router       /polls/{poll_id} [delete]
func (h *PollHandler) DeletePoll(c *gin.Context) {
	pollID, _ := strconv.Atoi(c.Param("poll_id"))
	userID := c.GetInt("user_id")

	if err := h.svc.DeletePoll(c.Request.Context(), userID, pollID); err != nil {
		code := http.StatusInternalServerError
		if err.Error() == "permission denied" {
			code = http.StatusForbidden
		}
		c.JSON(code, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Vote
// @Summary      Проголосовать
// @Tags         Polls
// @Security     BearerAuth
// @Param        poll_id  path  int      true  "ID опроса"
// @Param        input    body  dto.VoteReq  true "Тело запроса с данными"
// @Success      200  {object}  dto.ErrorResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /polls/{poll_id}/vote [post]
func (h *PollHandler) Vote(c *gin.Context) {
	// 1) Читаем pollID из пути
	pollID, err := strconv.Atoi(c.Param("poll_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll_id"})
		return
	}
	userID := c.GetInt("user_id")

	var req dto.VoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// 2) Передаём и pollID, и optionID
	if err := h.svc.Vote(c.Request.Context(), userID, pollID, req.OptionID); err != nil {
		switch err.Error() {
		case "permission denied":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		case "not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "voted"})
}

// ListPollsHandler возвращает опросы чата с пагинацией.
// GET /chats/:id/polls?limit=10&offset=0
func (h *PollHandler) ListPollsHandler(c *gin.Context) {
	// 1) Парсим chatID
	chatID, err := strconv.Atoi(c.Param("id"))
	if err != nil || chatID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}
	// 2) Получаем userID из контекста
	userID := c.GetInt("user_id")

	// 3) Пагинация
	limit := 10
	if ls := c.Query("limit"); ls != "" {
		if v, err := strconv.Atoi(ls); err == nil && v > 0 {
			limit = v
		}
	}
	offset := 0
	if os := c.Query("offset"); os != "" {
		if v, err := strconv.Atoi(os); err == nil && v >= 0 {
			offset = v
		}
	}

	// 4) Вызов сервиса
	polls, err := h.svc.ListPolls(c.Request.Context(), userID, chatID, limit, offset)
	if err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, polls)
}

// UnvoteHandler снимает голос.
// DELETE /chats/:id/polls/:poll_id/vote
func (h *PollHandler) UnvoteHandler(c *gin.Context) {
	// pollID
	pollID, err := strconv.Atoi(c.Param("poll_id"))
	if err != nil || pollID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}
	// userID
	userID := c.GetInt("user_id")

	var req dto.VoteReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.OptionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option_id"})
		return
	}

	if err := h.svc.Unvote(c.Request.Context(), userID, pollID, req.OptionID); err != nil {
		if err.Error() == "permission denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "vote removed"})
}
