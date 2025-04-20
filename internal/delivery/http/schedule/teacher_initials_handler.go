package schedule

import (
	"EduSync/internal/delivery/http/schedule/dto"
	"github.com/gin-gonic/gin"
	"net/http"

	"EduSync/internal/service"
)

type TeacherInitialsHandler struct {
	svc service.TeacherInitialsService
}

func NewTeacherInitialsHandler(svc service.TeacherInitialsService) *TeacherInitialsHandler {
	return &TeacherInitialsHandler{svc: svc}
}

// ListHandler — GET /schedule/initials
func (h *TeacherInitialsHandler) ListHandler(c *gin.Context) {
	ctx := c.Request.Context()
	institutionID, exist := c.Get("institution_id")
	if !exist {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения инициалов"})
		return
	}
	initials, err := h.svc.List(ctx, institutionID.(int))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "не удалось получить инициалы"})
		return
	}
	out := make([]*dto.TeacherInitialDTO, len(initials))
	for i, ti := range initials {
		out[i] = &dto.TeacherInitialDTO{
			ID:       ti.ID,
			Initials: ti.Initials,
		}
	}
	c.JSON(http.StatusOK, out)
}
