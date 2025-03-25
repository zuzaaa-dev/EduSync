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

// GetGroupsByInstitutionID возвращает группы для указанного учреждения.
func (h *GroupHandler) GetGroupsByInstitutionID(c *gin.Context) {
	institutionIDStr := c.Query("institution_id")
	institutionID, err := strconv.Atoi(institutionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный institution_id"})
		return
	}

	groups, err := h.service.GetGroupsByInstitutionID(c.Request.Context(), institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// GetGroupByID возвращает группы для указанного учреждения.
func (h *GroupHandler) GetGroupByID(c *gin.Context) {
	institutionIDStr := c.Query("id")
	institutionID, err := strconv.Atoi(institutionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный institution_id"})
		return
	}

	groups, err := h.service.GetGroupById(c.Request.Context(), institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}
