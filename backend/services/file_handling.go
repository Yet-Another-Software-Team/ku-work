package services

import (
	"context"
	"ku-work/backend/model"
	filehandling "ku-work/backend/providers/file_handling"
	"ku-work/backend/repository"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

type FileService struct {
	provider filehandling.FileHandlingProvider
	repo     repository.FileRepository
}

// NewFileService constructs a FileService with an injected provider and repository.
func NewFileService(repo repository.FileRepository, p filehandling.FileHandlingProvider) *FileService {
	if repo == nil {
		panic("file repository cannot be nil")
	}
	if p == nil {
		panic("file handling provider must not be nil")
	}
	return &FileService{provider: p, repo: repo}
}

// RegisterGlobal registers the service's provider as the global provider and
// installs the model-level storage deletion hook that delegates deletion to the provider.
func (s *FileService) RegisterGlobal() {
	// Register provider in the provider registry
	filehandling.RegisterProvider(s.provider)

	// Install model-level deletion hook pointing back to this service's provider.
	model.SetStorageDeleteHook(func(ctx context.Context, fileID string) error {
		return s.provider.DeleteFile(ctx, s.repo, fileID)
	})
}

// SaveFile delegates saving the uploaded file to the configured provider.
// It returns the created file record or an error.
func (s *FileService) SaveFile(ctx *gin.Context, userId string, file *multipart.FileHeader, category model.FileCategory) (*model.File, error) {
	return s.provider.SaveFile(ctx, s.repo, userId, file, category)
}

// ServeFile delegates serving a file to the configured provider.
func (s *FileService) ServeFile(ctx *gin.Context) {
	s.provider.ServeFile(ctx, s.repo)
}

// DeleteFile delegates deletion to the configured provider.
func (s *FileService) DeleteFile(ctx context.Context, fileID string) error {
	return s.provider.DeleteFile(ctx, s.repo, fileID)
}
