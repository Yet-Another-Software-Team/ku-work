package filehandling

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"ku-work/backend/helper"
	"ku-work/backend/model"
	"ku-work/backend/repository"

	"github.com/gin-gonic/gin"
)

// GCSProvider implements FileHandlingProvider using Google Cloud Storage.
type GCSProvider struct {
	BucketName string
	client     *storage.Client
	ctx        context.Context
}

// NewGCSProvider creates a new GCSProvider.
// If credentialsJSONPath is provided, it will be used to authenticate. Otherwise, Application Default Credentials are used.
func NewGCSProvider(ctx context.Context, bucketName string, credentialsJSONPath string) (*GCSProvider, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name is required")
	}
	var (
		client *storage.Client
		err    error
	)
	if credentialsJSONPath != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(credentialsJSONPath))
	} else {
		client, err = storage.NewClient(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create gcs client: %w", err)
	}
	return &GCSProvider{
		BucketName: bucketName,
		client:     client,
		ctx:        ctx,
	}, nil
}

// SaveFile uploads a file to GCS and creates a corresponding database record.
func (p *GCSProvider) SaveFile(ctx *gin.Context, repo repository.FileRepository, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error) {
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

	// Create DB record first to use its ID as the GCS object name.
	fileRecord := &model.File{
		UserID:   userId,
		Category: fileCategory,
	}
	if err := repo.Save(fileRecord); err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Optionally clean image metadata
	toWrite := data
	if fileCategory == model.FileCategoryImage {
		if clean, _, cerr := helper.CleanImageMetadata(data); cerr == nil && len(clean) > 0 {
			toWrite = clean
		}
	}

	// Upload to GCS with a timeout.
	reqCtx := p.ctx
	if ctx != nil && ctx.Request != nil && ctx.Request.Context() != nil {
		reqCtx = ctx.Request.Context()
	}
	uploadCtx, cancel := context.WithTimeout(reqCtx, 2*time.Minute)
	defer cancel()

	w := p.client.Bucket(p.BucketName).Object(fileRecord.ID).NewWriter(uploadCtx)
	w.ContentType = http.DetectContentType(toWrite)

	if _, err := io.Copy(w, bytes.NewReader(toWrite)); err != nil {
		_ = w.Close()
		_ = repo.Delete(fileRecord) // Rollback DB record
		return nil, fmt.Errorf("failed to upload to gcs: %w", err)
	}
	if err := w.Close(); err != nil {
		_ = repo.Delete(fileRecord) // Rollback DB record
		return nil, fmt.Errorf("failed to finalize upload to gcs: %w", err)
	}

	return fileRecord, nil
}

// ServeFile streams a file from GCS to the response.
func (p *GCSProvider) ServeFile(ctx *gin.Context, repo repository.FileRepository) {
	fileID := ctx.Param("fileID")
	if fileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file id is required"})
		return
	}

	readCtx, cancel := context.WithTimeout(ctx.Request.Context(), 2*time.Minute)
	defer cancel()

	obj := p.client.Bucket(p.BucketName).Object(fileID)
	r, err := obj.NewReader(readCtx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file from storage"})
		}
		return
	}
	defer func() { _ = r.Close() }()

	// Set content-type header if available, otherwise detect it.
	if r.Attrs.ContentType != "" {
		ctx.Header("Content-Type", r.Attrs.ContentType)
	} else {
		buffer := make([]byte, 512)
		n, _ := r.Read(buffer)
		if n > 0 {
			ctx.Header("Content-Type", http.DetectContentType(buffer[:n]))
			if _, wErr := ctx.Writer.Write(buffer[:n]); wErr != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write file to response"})
				return
			}
		}
	}

	// Stream the rest of the object content.
	if _, err := io.Copy(ctx.Writer, r); err != nil {
		log.Printf("Failed to copy file to response: %v", err)
	}
}

// DeleteFile deletes a file from GCS. It is idempotent, so it will not return an error if the file does not exist.
func (p *GCSProvider) DeleteFile(ctx context.Context, repo repository.FileRepository, fileID string) error {
	if fileID == "" {
		return fmt.Errorf("file id is required")
	}

	delCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	obj := p.client.Bucket(p.BucketName).Object(fileID)
	if err := obj.Delete(delCtx); err != nil && err != storage.ErrObjectNotExist {
		return fmt.Errorf("failed to delete object from gcs: %w", err)
	}

	// Delete DB record (idempotent behavior expected from repository)
	if repo != nil {
		if err := repo.Delete(&model.File{ID: fileID}); err != nil {
			return fmt.Errorf("failed to delete file record: %w", err)
		}
	}

	return nil
}
