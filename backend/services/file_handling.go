package services

import (
	"context"
	"fmt"
	"ku-work/backend/model"
	filehandling "ku-work/backend/services/file_handling"
	"mime/multipart"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileService struct {
	provider filehandling.FileHandlingProvider
	db       *gorm.DB
}

// NewFileService constructs a FileService by reading configuration from the environment.
// It supports 'local' and 'gcs' providers, configured via the FILE_PROVIDER environment variable.
// This function will panic if the configuration is invalid.
func NewFileService(db *gorm.DB) *FileService {
	providerType := strings.ToLower(strings.TrimSpace(os.Getenv("FILE_PROVIDER")))
	if providerType == "" || providerType == "local" {
		baseDir := os.Getenv("LOCAL_FILES_DIR")
		if strings.TrimSpace(baseDir) == "" {
			baseDir = "./files"
		}
		// Ensure directory exists
		// #nosec G301
		if err := os.MkdirAll(baseDir, 0750); err != nil {
			panic(fmt.Errorf("failed to create local files directory %s: %w", baseDir, err))
		}
		p := filehandling.NewLocalProvider(baseDir)
		return &FileService{provider: p, db: db}
	}
	if providerType == "gcs" {
		bucket := os.Getenv("GCS_BUCKET")
		if strings.TrimSpace(bucket) == "" {
			panic("GCS_BUCKET is required for gcs provider")
		}
		creds := os.Getenv("GCS_CREDENTIALS_PATH")
		p, err := filehandling.NewGCSProvider(context.Background(), bucket, creds)
		if err != nil {
			panic(fmt.Errorf("failed to create gcs provider: %w", err))
		}
		return &FileService{provider: p, db: db}
	}
	panic(fmt.Errorf("unsupported FILE_PROVIDER: %s", providerType))
}

// RegisterGlobal registers the service's provider as the global provider and
// installs the model-level storage deletion hook that delegates deletion to the provider.
func (s *FileService) RegisterGlobal() {
	// Register provider in the provider registry
	filehandling.RegisterProvider(s.provider)

	// Install model-level deletion hook pointing back to this service's provider.
	model.SetStorageDeleteHook(func(ctx context.Context, fileID string) error {
		return s.provider.DeleteFile(ctx, fileID)
	})
}

// SaveFile delegates saving the uploaded file to the configured provider.
// It returns the created file record or an error.
func (s *FileService) SaveFile(ctx *gin.Context, db *gorm.DB, userId string, file *multipart.FileHeader, category model.FileCategory) (*model.File, error) {
	tx := s.db
	if db != nil {
		tx = db
	}
	return s.provider.SaveFile(ctx, tx, userId, file, category)
}

// ServeFile delegates serving a file to the configured provider.
func (s *FileService) ServeFile(ctx *gin.Context) {
	s.provider.ServeFile(ctx, s.db)
}

// DeleteFile delegates deletion to the configured provider.
func (s *FileService) DeleteFile(ctx context.Context, fileID string) error {
	return s.provider.DeleteFile(ctx, fileID)
}
