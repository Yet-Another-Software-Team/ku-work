package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"ku-work/backend/model"
	"ku-work/backend/services"
	"mime/multipart"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	MAX_DOCS_SIZE    = 10 * 1024 * 1024 // 10MB
	MAX_IMAGE_SIZE   = 5 * 1024 * 1024  // 5MB
	MAX_IMAGE_WIDTH  = 4096             // 4096 pixels
	MAX_IMAGE_HEIGHT = 4096             // 4096 pixels
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

// SaveFile saves a file using the configured storage provider.
func SaveFile(ctx *gin.Context, db *gorm.DB, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error) {
	// Validate file category
	isValidCategory := fileCategory == model.FileCategoryImage || fileCategory == model.FileCategoryDocument
	if !isValidCategory {
		return nil, fmt.Errorf("invalid file category: %s", fileCategory)
	}

	if fileService == nil {
		return nil, fmt.Errorf("file service not configured")
	}

	// Delegate saving to the file service.
	return fileService.SaveFile(ctx, userId, file, fileCategory)
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
