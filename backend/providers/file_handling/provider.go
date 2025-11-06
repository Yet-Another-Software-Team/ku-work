package filehandling

import (
	"context"
	"errors"
	"mime/multipart"
	"sync"

	"ku-work/backend/model"
	"ku-work/backend/repository"

	"github.com/gin-gonic/gin"
)

type FileHandlingProvider interface {
	SaveFile(ctx *gin.Context, repo repository.FileRepository, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error)
	ServeFile(ctx *gin.Context, repo repository.FileRepository)
	DeleteFile(ctx context.Context, repo repository.FileRepository, fileID string) error
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
