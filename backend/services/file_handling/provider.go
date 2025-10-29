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

// registry is a package-level variable that holds the registered FileHandlingProvider.
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

// ProviderExists checks if a FileHandlingProvider is registered.
func ProviderExists() bool {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return registry != nil
}
