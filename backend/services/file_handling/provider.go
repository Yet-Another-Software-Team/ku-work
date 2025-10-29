package filehandling

import (
	"context"
	"errors"
	"mime/multipart"
	"sync"

	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileHandlingProvider interface {
	SaveFile(ctx *gin.Context, db *gorm.DB, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error)
	ServeFile(ctx *gin.Context, db *gorm.DB)
	DeleteFile(ctx context.Context, fileID string) error
}

// We maintain a small global registry so the application can register exactly one
// FileHandlingProvider at startup and other packages may obtain it without importing
// concrete implementation types. This simplifies wiring in handlers/models that need
// to coordinate file deletions when related DB records are removed.
var (
	registryMu sync.RWMutex
	registry   FileHandlingProvider
)

// RegisterProvider registers a FileHandlingProvider instance.
func RegisterProvider(p FileHandlingProvider) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = p
}

// GetProvider returns the currently registered FileHandlingProvider.
func GetProvider() (FileHandlingProvider, error) {
	registryMu.RLock()
	p := registry
	registryMu.RUnlock()
	if p == nil {
		return nil, errors.New("no file handling provider registered")
	}
	return p, nil
}

// MustGetProvider returns the currently registered FileHandlingProvider or panics.
func MustGetProvider() FileHandlingProvider {
	p, err := GetProvider()
	if err != nil {
		panic(err)
	}
	return p
}

// DeleteStoredFile deletes a file from the storage backend.
func DeleteStoredFile(ctx context.Context, fileID string) error {
	p, err := GetProvider()
	if err != nil {
		return err
	}
	return p.DeleteFile(ctx, fileID)
}

// DeleteStoredFileWithDB deletes a file from the storage backend and the database.
func DeleteStoredFileWithDB(ctx context.Context, db *gorm.DB, fileID string) error {
	_ = db // currently unused
	return DeleteStoredFile(ctx, fileID)
}

// ProviderExists checks if a FileHandlingProvider is registered.
func ProviderExists() bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry != nil
}
