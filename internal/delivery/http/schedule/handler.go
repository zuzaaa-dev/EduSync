package schedule

import (
	"EduSync/internal/delivery/http/schedule/dto"
	"EduSync/internal/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ScheduleHandler обрабатывает HTTP-запросы, связанные с расписанием.
// swagger:model
type ScheduleHandler struct {
	scheduleService service.ScheduleService
}

// NewScheduleHandler создает новый ScheduleHandler.
func NewScheduleHandler(scheduleService service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{scheduleService: scheduleService}
}

// UpdateScheduleHandler запускает обновление расписания для заданной группы.
// @Summary      Обновить расписание группы
// @Description  Запускает процесс обновления расписания для указанной группы
// @Tags         Schedule
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        group_name  query  string  true  "Название группы"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /schedule/update [get]
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
// @Summary      Получить расписание группы
// @Description  Возвращает полное расписание для указанной группы
// @Tags         Schedule
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        group_id  query  int  true  "ID группы"
// @Success      200  {array}   Item
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /schedule [get]
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

// swagger:route PUT /schedule/{id} schedule updateScheduleEntry
// @Summary      Обновить запись расписания
// @Description  Обновляет существующую запись в расписании (только для преподавателей)
// @Tags         Schedule
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  int  true  "ID записи расписания"
// @Param        input body  dto.UpdateScheduleReq  true  "Данные для обновления"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      403  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /schedule/{id} [put]
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

	if err := h.scheduleService.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления записи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "расписание обновлено"})
}

// swagger:route DELETE /schedule/{id} schedule deleteScheduleEntry
// DeleteHandler удаляет запись расписания
// @Summary      Удалить запись расписания
// @Description  Удаляет существующую запись из расписания (только для преподавателей)
// @Tags         Schedule
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id    path  int  true  "ID записи расписания"
// @Success      200  {object}  object{message=string}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      403  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /schedule/{id} [delete]
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

	if err := h.scheduleService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления записи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "запись удалена"})
}

// GetByTeacherInitialsHandler возвращает расписание по id инициалов преподавателя
// @Summary      Получить расписание преподавателя
// @Description  Возвращает расписание по идентификатору инициалов преподавателя
// @Tags         Schedule
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        initials_id  path  int  true  "ID инициалов преподавателя"
// @Success      200  {array}   Item
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /schedule/teacher_initials/{initials_id} [get]
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

// swagger:route POST /schedule schedule createScheduleEntry
// @Summary      Добавить запись в расписание
// @Description  Создаёт новую запись расписания (только для преподавателей)
// @Tags         Schedule
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input body dto.CreateScheduleReq true "Данные расписания"
// @Success      201 {object} object{message=string,id=int}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      403 {object} dto.ErrorResponse
// @Failure      409 {object} dto.ErrorResponse "Нарушение уникальности пары"
// @Failure      500 {object} dto.ErrorResponse
// @Router       /schedule [post]
func (h *ScheduleHandler) CreateHandler(c *gin.Context) {
	isTeacher, ok := c.Get("is_teacher")
	if !ok || !isTeacher.(bool) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Доступ разрешён только преподавателям"})
		return
	}

	var req dto.CreateScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	id, err := h.scheduleService.Create(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Пара уже занята"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка создания записи"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "расписание добавлено", "id": id})
}
