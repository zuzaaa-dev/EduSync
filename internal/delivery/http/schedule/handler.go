package schedule

import (
	"net/http"
	"strconv"

	service "EduSync/internal/service"
	"github.com/gin-gonic/gin"
)

// ScheduleHandler обрабатывает HTTP-запросы, связанные с расписанием.
type ScheduleHandler struct {
	scheduleService service.ScheduleService
}

// NewScheduleHandler создает новый ScheduleHandler.
func NewScheduleHandler(scheduleService service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

// UpdateScheduleHandler запускает обновление расписания для заданной группы.
// Например, вызов GET /schedule/update?group_id=1&group_name=ИС-38
func (h *ScheduleHandler) UpdateScheduleHandler(c *gin.Context) {
	groupName := c.Query("group_name")
	if groupName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_name обязательны"})
		return
	}

	err := h.scheduleService.Update(c.Request.Context(), groupName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Расписание успешно обновлено"})
}

// GetScheduleHandler возвращает расписание для заданной группы.
// Например, GET /schedule?group_id=1
func (h *ScheduleHandler) GetScheduleHandler(c *gin.Context) {
	groupIDStr := c.Query("group_id")
	if groupIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id обязателен"})
		return
	}
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный group_id"})
		return
	}

	scheduleEntries, err := h.scheduleService.ByGroupID(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, scheduleEntries)
}
