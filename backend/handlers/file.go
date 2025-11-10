package handlers

import (
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
)

type FileHandlers struct {
	fileService *services.FileService
}

func NewFileHandlers(s *services.FileService) *FileHandlers {
	return &FileHandlers{
		fileService: s,
	}
}

// @Summary Get a file
// @Description Serves a file from the server's file system using its unique ID. This is a public endpoint.
// @Tags Files
// @Produce octet-stream
// @Param fileID path string true "File ID"
// @Success 200 {file} file "The requested file"
// @Failure 400 {object} object{error=string} "Bad Request: Invalid file identifier or path"
// @Failure 404 {object} object{error=string} "Not Found: File not found"
// @Failure 500 {object} object{error=string} "Internal Server Error"
// @Router /files/{fileID} [get]
func (h *FileHandlers) ServeFileHandler(ctx *gin.Context) {
	if h.fileService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "file service not configured"})
		return
	}
	h.fileService.ServeFile(ctx)
}
