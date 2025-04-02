package subject

import (
	"EduSync/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// InstitutionHandler обрабатывает HTTP-запросы, связанные с предметами.
type InstitutionHandler struct {
	subjectService service.SubjectService
}

// NewInstitutionHandler создаёт новый экземпляр хэндлера предметов.
func NewInstitutionHandler(subjectService service.SubjectService) *InstitutionHandler {
	return &InstitutionHandler{subjectService: subjectService}
}

// CreateSubjectHandler обрабатывает создание предмета.
func (h *InstitutionHandler) CreateSubjectHandler(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required"`
		InstitutionID int    `json:"institution_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	subjectID, err := h.subjectService.Create(c.Request.Context(), req.Name, req.InstitutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"subject_id": subjectID})
}

// GetSubjectsByInstitution обрабатывает получение списка предметов по ID учебного заведения.
func (h *InstitutionHandler) GetSubjectsByInstitution(c *gin.Context) {
	institutionID, err := strconv.Atoi(c.Param("institution_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неправильный аргумент"})
		return
	}
	if institutionID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неправильный аргумент"})
		return
	}
	subjects, err := h.subjectService.ByInstitutionID(c.Request.Context(), institutionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if subjects == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "предметов нет"})
		return
	}
	c.JSON(http.StatusOK, subjects)
}

// GetSubjectsByGroup обрабатывает получение списка предметов по ID группы.
func (h *InstitutionHandler) GetSubjectsByGroup(c *gin.Context) {
	groupID, err := strconv.Atoi(c.Param("group_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неправильный аргумент"})
		return
	}
	if groupID < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неправильный аргумент"})
		return
	}
	subjects, err := h.subjectService.ByGroupID(c.Request.Context(), groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if subjects == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "предметов нет"})
		return
	}
	c.JSON(http.StatusOK, subjects)
}
