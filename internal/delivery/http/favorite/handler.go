package favorite

import (
	"EduSync/internal/domain/chat"
	"EduSync/internal/service"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type FileFavoriteHandler struct {
	svc service.FileFavoriteService
}

func NewFileFavoriteHandler(svc service.FileFavoriteService) *FileFavoriteHandler {
	return &FileFavoriteHandler{svc: svc}
}

// AddFavoriteFile
// @Summary      Добавить файл в избранное
// @Description  Помечает файл как избранный для текущего пользователя
// @Tags         Files
// @Security     BearerAuth
// @Param        id   path      int  true  "ID файла"
// @Success      204
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      409  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /files/{id}/favorite [post]
func (h *FileFavoriteHandler) AddFavoriteFile(c *gin.Context) {
	fileID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}
	userID := c.GetInt("user_id")

	err = h.svc.AddFavorite(c.Request.Context(), userID, fileID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	case chat.ErrNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case chat.ErrPermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	case chat.ErrAlreadyFavorited:
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

// RemoveFavoriteFile
// @Summary      Удалить файл из избранного
// @Description  Снимает пометку «избранный» с файла
// @Tags         Files
// @Security     BearerAuth
// @Param        id   path      int  true  "ID файла"
// @Success      204
// @Failure      403  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      409  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /files/{id}/favorite [delete]
func (h *FileFavoriteHandler) RemoveFavoriteFile(c *gin.Context) {
	fileID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}
	userID := c.GetInt("user_id")

	err = h.svc.RemoveFavorite(c.Request.Context(), userID, fileID)
	switch err {
	case nil:
		c.Status(http.StatusNoContent)
	case chat.ErrNotFound, chat.ErrNotFavorited:
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case chat.ErrPermissionDenied:
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}

// ListFavoriteFiles
// @Summary      Список избранных файлов
// @Description  Возвращает для текущего пользователя список избранных файлов (id + url)
// @Tags         Files
// @Security     BearerAuth
// @Produce      json
// @Success      200  {array}   dto.FileInfo
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /files/favorites [get]
func (h *FileFavoriteHandler) ListFavoriteFiles(c *gin.Context) {
	userID := c.GetInt("user_id")

	files, err := h.svc.ListFavorites(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, files)
}
