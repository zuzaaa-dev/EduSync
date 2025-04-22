package institution

import (
	"net/http"
	"strconv"

	service "EduSync/internal/service"
	"github.com/gin-gonic/gin"
)

// InstitutionHandler обрабатывает HTTP-запросы, связанные с учебными заведениями.
type InstitutionHandler struct {
	instService      service.InstitutionService
	emailMaskService service.EmailMaskService
}

// NewInstitutionHandler создает новый обработчик.
func NewInstitutionHandler(instService service.InstitutionService, emailMaskService service.EmailMaskService) *InstitutionHandler {
	return &InstitutionHandler{
		instService:      instService,
		emailMaskService: emailMaskService,
	}
}

// GetInstitutionByID возвращает учреждение по ID
// @Summary      Получить учреждение
// @Description  Возвращает информацию об учебном заведении по его ID
// @Tags         Institutions
// @Accept       json
// @Produce      json
// @Param        id  path  int  true  "ID учреждения"
// @Success      200  {object}  Institution
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /institutions/{id} [get]
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

// GetAllInstitutions возвращает все учреждения
// @Summary      Список учреждений
// @Description  Возвращает список всех учебных заведений
// @Tags         Institutions
// @Accept       json
// @Produce      json
// @Success      200  {array}   Institution
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /institutions [get]
func (h *InstitutionHandler) GetAllInstitutions(c *gin.Context) {
	insts, err := h.instService.All(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, insts)
}

// GetAllMasks возвращает все почтовые маски
// @Summary      Список почтовых масок
// @Description  Возвращает список всех зарегистрированных почтовых масок
// @Tags         Institutions
// @Accept       json
// @Produce      json
// @Success      200  {array}   EmailMask
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /email-masks [get]
func (h *InstitutionHandler) GetAllMasks(c *gin.Context) {
	masks, err := h.emailMaskService.AllMasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения почтовых масок"})
		return
	}
	c.JSON(http.StatusOK, masks)
}

// GetMaskByValue возвращает маску по значению
// @Summary      Найти маску по значению
// @Description  Возвращает информацию о почтовой маске по ее значению
// @Tags         Institutions
// @Accept       json
// @Produce      json
// @Param        mask  path  string  true  "Значение маски"
// @Success      200  {object}  EmailMask
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /email-masks/{mask} [get]
func (h *InstitutionHandler) GetMaskByValue(c *gin.Context) {
	maskValue := c.Param("mask")
	if maskValue == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Параметр mask обязателен"})
		return
	}
	mask, err := h.emailMaskService.MaskByValue(c.Request.Context(), maskValue)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mask)
}
