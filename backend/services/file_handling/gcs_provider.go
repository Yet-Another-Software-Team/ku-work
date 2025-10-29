package filehandling

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"ku-work/backend/helper"
	"ku-work/backend/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GCSProvider implements FileHandlingProvider using Google Cloud Storage.
//
// This file is enabled only when the `gcs` build tag is set.
type GCSProvider struct {
	BucketName string
	client     *storage.Client
	// Optionally store a default context for operations that do not have a request context.
	ctx context.Context
}

// NewGCSProvider creates a GCSProvider. If credentialsJSONPath is non-empty it will be
// used to configure the client via option.WithCredentialsFile. The provided ctx is used
// as a base context for client creation and long-lived operations.
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

// SaveFile uploads the provided file to GCS and creates the DB record. The DB record's ID
// is used as the storage object name/key.
func (p *GCSProvider) SaveFile(ctx *gin.Context, db *gorm.DB, userId string, file *multipart.FileHeader, fileCategory model.FileCategory) (*model.File, error) {
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

	// Create DB record first so we have an ID for the object name
	fileRecord := &model.File{
		UserID:   userId,
		Category: fileCategory,
	}
	if err := db.Create(fileRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create file record: %w", err)
	}

	// Optionally clean image metadata
	toWrite := data
	if fileCategory == model.FileCategoryImage {
		if clean, _, cerr := helper.CleanImageMetadata(data); cerr == nil && len(clean) > 0 {
			toWrite = clean
		}
	}

	// Upload to GCS. Use request context if present to bound the operation.
	reqCtx := p.ctx
	if ctx != nil && ctx.Request != nil && ctx.Request.Context() != nil {
		reqCtx = ctx.Request.Context()
	}
	// Add a timeout so uploads don't hang indefinitely.
	uploadCtx, cancel := context.WithTimeout(reqCtx, 2*time.Minute)
	defer cancel()

	w := p.client.Bucket(p.BucketName).Object(fileRecord.ID).NewWriter(uploadCtx)
	w.ContentType = http.DetectContentType(toWrite)
	// You can set other attributes here (CacheControl, ACLs, etc) if desired.

	if _, err := io.Copy(w, bytes.NewReader(toWrite)); err != nil {
		_ = w.Close()
		_ = db.Delete(fileRecord).Error
		return nil, fmt.Errorf("failed to upload to gcs: %w", err)
	}
	if err := w.Close(); err != nil {
		_ = db.Delete(fileRecord).Error
		return nil, fmt.Errorf("failed to finalize upload to gcs: %w", err)
	}

	return fileRecord, nil
}

// ServeFile streams a file from GCS to the response. It uses the file ID from the route
// parameters (param name: "fileID").
func (p *GCSProvider) ServeFile(ctx *gin.Context, db *gorm.DB) {
	fileID := ctx.Param("fileID")
	// Basic validation to avoid malformed keys (GCS allows many characters, but we avoid obvious attacks)
	if fileID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "file id is required"})
		return
	}

	reqCtx := p.ctx
	if ctx != nil && ctx.Request != nil && ctx.Request.Context() != nil {
		reqCtx = ctx.Request.Context()
	}
	readCtx, cancel := context.WithTimeout(reqCtx, 2*time.Minute)
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

	// Set content-type header if available
	if r.Attrs.ContentType != "" {
		ctx.Header("Content-Type", r.Attrs.ContentType)
	} else {
		// Fallback: read a small chunk to detect content type
		buffer := make([]byte, 512)
		n, _ := r.Read(buffer)
		if n > 0 {
			ctx.Header("Content-Type", http.DetectContentType(buffer[:n]))
			// write the first chunk
			if _, wErr := ctx.Writer.Write(buffer[:n]); wErr != nil {
				return
			}
		}
		// Stream remainder
		if _, err := io.Copy(ctx.Writer, r); err != nil {
			// nothing sensible to do after headers/body started
		}
		return
	}

	// Stream the object content
	if _, err := io.Copy(ctx.Writer, r); err != nil {
		// If streaming fails after response started, there's little we can do.
		return
	}
}

// DeleteFile deletes the object identified by fileID from the configured GCS bucket.
// Behavior:
// - If the object does not exist, return nil (idempotent).
// - If fileID is empty, return an error.
func (p *GCSProvider) DeleteFile(ctx context.Context, fileID string) error {
	if fileID == "" {
		return fmt.Errorf("file id is required")
	}

	// Choose a base context: prefer the provided ctx; fall back to provider context or background.
	reqCtx := ctx
	if reqCtx == nil {
		if p.ctx != nil {
			reqCtx = p.ctx
		} else {
			reqCtx = context.Background()
		}
	}

	// Apply a sensible timeout for delete operations.
	delCtx, cancel := context.WithTimeout(reqCtx, 1*time.Minute)
	defer cancel()

	obj := p.client.Bucket(p.BucketName).Object(fileID)
	if err := obj.Delete(delCtx); err != nil {
		// Treat 'not found' as success to make deletion idempotent.
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("failed to delete object from gcs: %w", err)
	}
	return nil
}
