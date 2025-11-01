package handlers

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"

	"github.com/chai2010/webp"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	"log"
	"mime/multipart"
	"os"

	"path/filepath"
	"strings"

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

// SaveFile saves a file to disk and creates a File record in the database
// Returns the file ID for referencing in other models
func SaveFile(ctx *gin.Context, db *gorm.DB, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error) {
	// Validate file category
	isValidCategory := fileCategory == model.FileCategoryImage || fileCategory == model.FileCategoryDocument
	if !isValidCategory {
		return nil, fmt.Errorf("invalid file category: %s", fileCategory)
	}

	fileType, err := ExtractFileType(file.Filename)
	if err != nil {
		return nil, fmt.Errorf("unsupported file type: %s", err)
	}

	// Check Docs file size
	if fileCategory == model.FileCategoryDocument && file.Size > MAX_DOCS_SIZE {
		return nil, fmt.Errorf("file size exceeds limit")
	} else if fileCategory == model.FileCategoryImage {
		// Check Image file size
		if file.Size > MAX_IMAGE_SIZE {
			return nil, fmt.Errorf("image file size exceeds limit of 5MB")
		}

		// Check Image dimensions
		src, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("could not open image file to check dimensions: %w", err)
		}
		defer func() { _ = src.Close() }()

		config, _, err := image.DecodeConfig(src)
		if err != nil {
			return nil, fmt.Errorf("invalid image file: could not read dimensions")
		}

		if config.Width > MAX_IMAGE_WIDTH || config.Height > MAX_IMAGE_HEIGHT {
			return nil, fmt.Errorf("image dimensions exceed the maximum limit of %dx%d pixels", MAX_IMAGE_WIDTH, MAX_IMAGE_HEIGHT)
		}
	}

	fileRecord := &model.File{
		UserID:   userId,
		FileType: fileType,
		Category: fileCategory,
	}

	if err := db.Create(fileRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create file record: %s", err)
	}

	filePath := filepath.Join("./files", fileRecord.ID)

	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		db.Delete(fileRecord)
		return nil, fmt.Errorf("failed to save file to disk: %s", err)
	}

	if fileCategory == model.FileCategoryImage {
		// Clean image metadata
		if err := CleanImageMetadata(filePath); err != nil {
			log.Printf("Failed to clean image metadata: %s", err)
		}
	}

	return fileRecord, nil
}

// Get file type from file name
func ExtractFileType(filename string) (model.FileType, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "", fmt.Errorf("no file extension found")
	}

	ext = ext[1:]

	switch ext {
	case "jpg":
		return model.FileTypeJPG, nil
	case "jpeg":
		return model.FileTypeJPEG, nil
	case "png":
		return model.FileTypePNG, nil
	case "webp":
		return model.FileTypeWEBP, nil
	case "pdf":
		return model.FileTypePDF, nil
	case "doc":
		return model.FileTypeDOC, nil
	case "docx":
		return model.FileTypeDOCX, nil
	default:
		return "", fmt.Errorf("unsupported file type: %s", ext)
	}
}

// Clean image metadata using file path
//
// Skip unsupported formats.
// Return error if any
func CleanImageMetadata(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}

	img, format, err := image.Decode(file)
	if err != nil {
		_ = file.Close()
		return fmt.Errorf("could not decode image: %w", err)
	}
	_ = file.Close()

	isSupported := format == "jpeg" || format == "png" || format == "webp"
	if !isSupported {
		log.Printf("Metadata cleaning not supported for format '%s', skipping file: %s", format, filePath)
		return nil
	}

	// Create a temporary file in the same directory to write the cleaned image.
	tempFile, err := os.CreateTemp(filepath.Dir(filePath), "clean-img-")
	if err != nil {
		return fmt.Errorf("could not create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tempFile.Name()) }() // Ensure temp file is cleaned up on error

	switch format {
	case "jpeg":
		if err := jpeg.Encode(tempFile, img, nil); err != nil {
			_ = tempFile.Close()
			return fmt.Errorf("could not encode jpeg: %w", err)
		}
	case "png":
		if err := png.Encode(tempFile, img); err != nil {
			_ = tempFile.Close()
			return fmt.Errorf("could not encode png: %w", err)
		}
	case "webp":
		if err := webp.Encode(tempFile, img, &webp.Options{Lossless: false}); err != nil {
			_ = tempFile.Close()
			return fmt.Errorf("could not encode webp: %w", err)
		}
	}

	// Close the temp file before renaming.
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("could not close temp file: %w", err)
	}

	// Atomically replace the original file with the new, cleaned file.
	if err := os.Rename(tempFile.Name(), filePath); err != nil {
		return fmt.Errorf("could not replace original file with cleaned version: %w", err)
	}

	return nil
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
	fileID := ctx.Param("fileID")

	// Ensure that file id not contain invalid characters
	if strings.Contains(fileID, "/") || strings.Contains(fileID, `\`) || strings.Contains(fileID, "..") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file identifier"})
		return
	}

	// Check file deactivation from DB
	fileDBO := &model.File{
		ID: fileID,
	}

	if err := h.DB.First(fileDBO).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	if helper.IsDeactivated(h.DB, fileDBO.UserID) {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "File is deactivated"})
		return
	}

	// Get absolute path of base directory
	// TODO: Externalize Absolute Path directory to allow usage of non-default base directories
	baseDir := "./files"
	filePath := filepath.Join(baseDir, fileID)
	absBaseDir, err := filepath.Abs(baseDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}
	// Ensure that absFilePath is inside absBaseDir
	if !strings.HasPrefix(absFilePath, absBaseDir+string(os.PathSeparator)) && absFilePath != absBaseDir {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File path traversal detected"})
		return
	}

	file, err := os.Open(absFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		}
		return
	}
	// Close file on function exit
	defer func() { _ = file.Close() }()

	//Read file content to buffer
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	mimeType := http.DetectContentType(buffer[:n])
	ctx.Header("Content-Type", mimeType)

	ctx.File(absFilePath)
}
