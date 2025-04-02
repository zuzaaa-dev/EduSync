package institution

import (
	"net/http"
	"strconv"

	service "EduSync/internal/service"
	"github.com/gin-gonic/gin"
)

// InstitutionHandler обрабатывает HTTP-запросы, связанные с учебными заведениями.
type InstitutionHandler struct {
	instService service.InstitutionService
}

func NewInstitutionHandler(s service.InstitutionService) *InstitutionHandler {
	return &InstitutionHandler{instService: s}
}

// GetInstitutionByID возвращает учреждение по заданному id.
func (h *InstitutionHandler) GetInstitutionByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный id учреждения"})
		return
	}

	inst, err := h.instService.ByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if inst == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Учреждение не найдено"})
		return
	}
	c.JSON(http.StatusOK, inst)
}

// GetAllInstitutions возвращает список всех учреждений.
func (h *InstitutionHandler) GetAllInstitutions(c *gin.Context) {
	insts, err := h.instService.All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, insts)
}
