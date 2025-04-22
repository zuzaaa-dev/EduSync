package group

import (
	"EduSync/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GroupHandler обрабатывает HTTP-запросы, связанные с группами.
type GroupHandler struct {
	service service.GroupService
}

// NewGroupHandler создает новый GroupHandler.
func NewGroupHandler(service service.GroupService) *GroupHandler {
	return &GroupHandler{service: service}
}

// GetGroupsByInstitutionID возвращает группы учреждения
// @Summary      Получить группы учреждения
// @Description  Возвращает список всех групп для указанного учебного заведения
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        institution_id  path  int  true  "ID учебного заведения"
// @Success      200  {array}   Group
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /institutions/{institution_id}/groups [get]
func (h *GroupHandler) GetGroupsByInstitutionID(c *gin.Context) {
	institutionIDStr := c.Param("institution_id")
	institutionID, err := strconv.Atoi(institutionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный institution_id"})
		return
	}

	groups, err := h.service.ByInstitutionID(c.Request.Context(), institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// GetGroupByID возвращает группу по ID
// @Summary      Получить группу по ID
// @Description  Возвращает информацию о конкретной группе
// @Tags         Groups
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "ID группы"
// @Success      200  {object}  Group
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /groups/{id} [get]
func (h *GroupHandler) GetGroupByID(c *gin.Context) {
	institutionIDStr := c.Param("id")
	institutionID, err := strconv.Atoi(institutionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный institution_id"})
		return
	}

	groups, err := h.service.ById(c.Request.Context(), institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}
