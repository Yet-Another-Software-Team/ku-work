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
	// BaseDir is the directory where files are stored.
	BaseDir string
}

// NewLocalProvider constructs a LocalProvider.
func NewLocalProvider(baseDir string) *LocalProvider {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = "./files"
	}
	return &LocalProvider{BaseDir: baseDir}
}

// SaveFile saves a file to the local filesystem.
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

	// Create DB record first to use its ID as the filename.
	fileRecord := &model.File{
		UserID:   userId,
		Category: fileCategory,
	}
	if err := db.Create(fileRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Ensure base directory exists
	if err := os.MkdirAll(p.BaseDir, 0o755); err != nil {
		_ = db.Delete(fileRecord).Error // Rollback DB record
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
		_ = db.Delete(fileRecord) // Rollback DB record
		return nil, fmt.Errorf("failed to write file to disk: %w", err)
	}

	return fileRecord, nil
}

// ServeFile serves a file from the local filesystem.
func (p *LocalProvider) ServeFile(ctx *gin.Context, db *gorm.DB) {
	fileID := ctx.Param("fileID")
	// Basic sanitization to avoid path traversal
	if strings.Contains(fileID, "/") || strings.Contains(fileID, `\`) || strings.Contains(fileID, "..") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file identifier"})
		return
	}

	filePath := filepath.Join(p.BaseDir, fileID)

	absBaseDir, err := filepath.Abs(p.BaseDir)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}
	// Prevent path traversal
	if !strings.HasPrefix(absFilePath, absBaseDir+string(os.PathSeparator)) && absFilePath != absBaseDir {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File path traversal detected"})
		return
	}

	ctx.File(absFilePath)
}

// DeleteFile removes a file from the local filesystem. It is idempotent.
func (p *LocalProvider) DeleteFile(ctx context.Context, fileID string) error {
	// Basic validation to avoid path traversal
	if strings.Contains(fileID, "/") || strings.Contains(fileID, `\`) || strings.Contains(fileID, "..") {
		return fmt.Errorf("invalid file identifier")
	}

	path := filepath.Join(p.BaseDir, fileID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove local file: %w", err)
	}
	return nil
}
