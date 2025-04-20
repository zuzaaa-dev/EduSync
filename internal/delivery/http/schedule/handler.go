package schedule

import (
	"EduSync/internal/delivery/http/schedule/dto"
	"EduSync/internal/service"
	"net/http"
	"strconv"

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

	err := h.scheduleService.Save(c.Request.Context(), groupName)
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

func (h *ScheduleHandler) UpdateHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID записи"})
		return
	}

	// Только преподаватель может обновлять
	isTeacher, ok := c.Get("is_teacher")
	if !ok || !isTeacher.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "доступ разрешён только преподавателям"})
		return
	}

	var req dto.UpdateScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	userID := c.GetInt("user_id")
	if err := h.scheduleService.Update(c.Request.Context(), userID, id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления записи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "расписание обновлено"})
}

func (h *ScheduleHandler) DeleteHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID записи"})
		return
	}

	// Только преподаватель может удалять
	isTeacher, ok := c.Get("is_teacher")
	if !ok || !isTeacher.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "доступ разрешён только преподавателям"})
		return
	}

	userID := c.GetInt("user_id")
	if err := h.scheduleService.Delete(c.Request.Context(), userID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления записи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "запись удалена"})
}

// GetByTeacherInitialsHandler возвращает расписание по id инициалов преподавателя.
// GET /schedule/teacher_initials/:initials_id
func (h *ScheduleHandler) GetByTeacherInitialsHandler(c *gin.Context) {
	idStr := c.Param("initials_id")
	initialsID, err := strconv.Atoi(idStr)
	if err != nil || initialsID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный идентификатор initials_id"})
		return
	}

	entries, err := h.scheduleService.ByTeacherInitialsID(c.Request.Context(), initialsID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить расписание"})
		return
	}
	c.JSON(http.StatusOK, entries)
}
