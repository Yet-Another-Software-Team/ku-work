package handlers

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"github.com/chai2010/webp"

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
	MAX_IMAGE_SIZE   = 1 * 1024 * 1024  // 1MB
	MAX_IMAGE_WIDTH  = 2048             // 2048 pixels
	MAX_IMAGE_HEIGHT = 2048             // 2048 pixels
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
func (h *FileHandlers) SaveFile(ctx *gin.Context, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (string, error) {
	// Validate file category
	isValidCategory := fileCategory == model.FileCategoryImage || fileCategory == model.FileCategoryDocument
	if !isValidCategory {
		return "", fmt.Errorf("invalid file category: %s", fileCategory)
	}

	fileType, err := ExtractFileType(file.Filename)
	if err != nil {
		return "", fmt.Errorf("unsupported file type: %s", err)
	}

	// Check Docs file size
	if fileCategory == model.FileCategoryDocument && file.Size > MAX_DOCS_SIZE {
		return "", fmt.Errorf("file size exceeds limit")
	} else 
	if fileCategory == model.FileCategoryImage {
		// Check Image file size
		if file.Size > MAX_IMAGE_SIZE {
			return "", fmt.Errorf("image file size exceeds limit of 1MB")
		}

		// Check Image dimensions
		src, err := file.Open()
		if err != nil {
			return "", fmt.Errorf("could not open image file to check dimensions: %w", err)
		}
		defer src.Close()

		config, _, err := image.DecodeConfig(src)
		if err != nil {
			return "", fmt.Errorf("invalid image file: could not read dimensions")
		}

		if config.Width > MAX_IMAGE_WIDTH || config.Height > MAX_IMAGE_HEIGHT {
			return "", fmt.Errorf("image dimensions exceed the maximum limit of %dx%d pixels", MAX_IMAGE_WIDTH, MAX_IMAGE_HEIGHT)
		}
	}

	fileRecord := model.File{
		UserID:   userId,
		FileType: fileType,
		Category: fileCategory,
	}

	if err := h.DB.Create(&fileRecord).Error; err != nil {
		return "", fmt.Errorf("failed to create file record: %s", err)
	}

	filePath := filepath.Join("./files", fileRecord.ID)

	
	if err := ctx.SaveUploadedFile(file, filePath); err != nil {
		h.DB.Delete(&fileRecord)
		return "", fmt.Errorf("failed to save file to disk: %s", err)
	}

	if fileCategory == model.FileCategoryImage {
		// Clean image metadata
		if err := CleanImageMetadata(filePath); err != nil {
			log.Printf("Failed to clean image metadata: %s", err)
		}
	}

	return fileRecord.ID, nil
}

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

func CleanImageMetadata(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}

	img, format, err := image.Decode(file)
	if err != nil {
		file.Close()
		return fmt.Errorf("could not decode image: %w", err)
	}
	file.Close()

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
	defer os.Remove(tempFile.Name()) // Ensure temp file is cleaned up on error

	switch format {
	case "jpeg":
		if err := jpeg.Encode(tempFile, img, nil); err != nil {
			tempFile.Close()
			return fmt.Errorf("could not encode jpeg: %w", err)
		}
	case "png":
		if err := png.Encode(tempFile, img); err != nil {
			tempFile.Close()
			return fmt.Errorf("could not encode png: %w", err)
		}
	case "webp":
		if err := webp.Encode(tempFile, img, &webp.Options{Lossless: true}); err != nil {
			tempFile.Close()
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


