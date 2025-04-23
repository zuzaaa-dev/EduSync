package material

import (
	chatSvc "EduSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type MaterialHandler struct {
	svc chatSvc.FileService
}

func NewFileHandler(svc chatSvc.FileService) *MaterialHandler {
	return &MaterialHandler{svc: svc}
}

// GetFileHandler отдаёт контент файла (streaming).
// GET /files/:id
func (h *MaterialHandler) GetFileHandler(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}
	userID := c.GetInt("user_id")

	reader, fname, err := h.svc.File(c.Request.Context(), userID, id)
	if err != nil {
		// если это permission denied или not found
		switch err.Error() {
		case "file not found":
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case "permission denied":
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		}
		return
	}
	defer reader.Close()
	c.File(fname)
}
