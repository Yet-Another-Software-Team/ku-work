package filehandling

import (
	"context"
	"fmt"
	"io"
	"ku-work/backend/helper"
	"ku-work/backend/model"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// LocalProvider stores files on the local filesystem.
type LocalProvider struct {
	// BaseDir is the directory where files are stored. Defaults to "./files" when empty.
	BaseDir string
}

// NewLocalProvider constructs a LocalProvider. If baseDir is empty it defaults to "./files".
func NewLocalProvider(baseDir string) *LocalProvider {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "./files"
	}
	return &LocalProvider{BaseDir: baseDir}
}

// SaveFile implements FileHandlingProvider for local filesystem storage.
func (p *LocalProvider) SaveFile(ctx *gin.Context, db *gorm.DB, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error) {
	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("could not open uploaded file: %w", err)
	}
	defer func() { _ = src.Close() }()

	// Read content
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("could not read uploaded file: %w", err)
	}

	// Validate using helper
	if ok, vErr := helper.IsValidFile(data, fileCategory); !ok {
		if vErr != nil {
			return nil, vErr
		}
		return nil, fmt.Errorf("file validation failed")
	}

	// Create DB record first so we have an ID for the stored object
	fileRecord := &model.File{
		UserID:   userId,
		Category: fileCategory,
	}
	if err := db.Create(fileRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Ensure base directory exists
	if err := os.MkdirAll(p.BaseDir, 0o755); err != nil {
		_ = db.Delete(fileRecord).Error
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Optionally clean image metadata
	toWrite := data
	if fileCategory == model.FileCategoryImage {
		if clean, _, cerr := helper.CleanImageMetadata(data); cerr == nil && len(clean) > 0 {
			toWrite = clean
		}
	}

	targetPath := filepath.Join(p.BaseDir, fileRecord.ID)
	if err := os.WriteFile(targetPath, toWrite, 0o644); err != nil {
		_ = db.Delete(fileRecord)
		return nil, fmt.Errorf("failed to write file to disk: %w", err)
	}

	return fileRecord, nil
}

// ServeFile implements FileHandlingProvider for local filesystem storage.
func (p *LocalProvider) ServeFile(ctx *gin.Context, db *gorm.DB) {
	fileID := ctx.Param("fileID")
	// basic sanitization to avoid path traversal
	if strings.Contains(fileID, "/") || strings.Contains(fileID, `\`) || strings.Contains(fileID, "..") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file identifier"})
		return
	}

	baseDir := p.BaseDir
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

	f, err := os.Open(absFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
		}
		return
	}
	defer func() { _ = f.Close() }()

	// Probe content type
	buffer := make([]byte, 512)
	n, rErr := f.Read(buffer)
	if rErr != nil && rErr != io.EOF {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	contentType := http.DetectContentType(buffer[:n])
	ctx.Header("Content-Type", contentType)

	// Use Gin's File convenience (it will re-open and serve the file)
	ctx.File(absFilePath)
}

// DeleteFile removes the file from local disk. It is idempotent: attempting to delete a
// non-existent file is not an error. The method uses the provided context for potential
// future cancellation/timeouts (currently there is no blocking IO that benefits from ctx).
func (p *LocalProvider) DeleteFile(ctx context.Context, fileID string) error {
	// Basic validation to avoid path traversal-like identifiers
	if strings.Contains(fileID, "/") || strings.Contains(fileID, `\`) || strings.Contains(fileID, "..") {
		return fmt.Errorf("invalid file identifier")
	}

	// Construct path and attempt removal
	path := filepath.Join(p.BaseDir, fileID)
	err := os.Remove(path)
	if err != nil {
		// Treat not-exist as success (idempotent)
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to remove local file: %w", err)
	}
	return nil
}
