package handlers

import (
	"net/http"

	"ku-work/backend/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileHandlers struct {
	DB *gorm.DB
}

func NewFileHandlers(db *gorm.DB) *FileHandlers {
	return &FileHandlers{
		DB: db,
	}
}

// fileService is the package-level FileService used by handlers.
// It must be set during application initialization.
var fileService *services.FileService

// SetFileService sets the FileService instance for the handlers.
func SetFileService(s *services.FileService) {
	fileService = s
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
	if fileService == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "file service not configured"})
		return
	}
	// Delegate serving to the file service.
	fileService.ServeFile(ctx)
}
