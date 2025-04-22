package subject

import (
	"EduSync/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SubjectHandler обрабатывает HTTP-запросы, связанные с предметами.
type SubjectHandler struct {
	subjectService service.SubjectService
}

// NewInstitutionHandler создаёт новый экземпляр хэндлера предметов.
func NewInstitutionHandler(subjectService service.SubjectService) *SubjectHandler {
	return &SubjectHandler{subjectService: subjectService}
}

// CreateSubjectHandler создает новый предмет
// @Summary      Создать предмет
// @Description  Создает новую запись учебного предмета
// @Tags         Subjects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input  body      object{name=string,institution_id=int}  true  "Данные предмета"
// @Success      201  {object}  object{subject_id=int}
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /subjects [post]
func (h *SubjectHandler) CreateSubjectHandler(c *gin.Context) {
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

// GetSubjectsByInstitution возвращает предметы учреждения
// @Summary      Получить предметы учреждения
// @Description  Возвращает список предметов для указанного учебного заведения
// @Tags         Subjects
// @Accept       json
// @Produce      json
// @Param        institution_id  path  int  true  "ID учебного заведения"
// @Success      200  {array}   Subject
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /subjects/institution/{institution_id} [get]
func (h *SubjectHandler) GetSubjectsByInstitution(c *gin.Context) {
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

// GetSubjectsByGroup возвращает предметы группы
// @Summary      Получить предметы группы
// @Description  Возвращает список предметов для указанной группы
// @Tags         Subjects
// @Accept       json
// @Produce      json
// @Param        group_id  path  int  true  "ID группы"
// @Success      200  {array}   Subject
// @Failure      400  {object} dto.ErrorResponse
// @Failure      500  {object} dto.ErrorResponse
// @Router       /subjects/group/{group_id} [get]
func (h *SubjectHandler) GetSubjectsByGroup(c *gin.Context) {
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
