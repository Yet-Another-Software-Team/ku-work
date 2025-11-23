package helper

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"ku-work/backend/model"
	"net/http"
	"strings"

	"github.com/chai2010/webp"
)

const (
	MAX_DOCS_SIZE    = 10 * 1024 * 1024 // 10MB
	MAX_IMAGE_SIZE   = 5 * 1024 * 1024  // 5MB
	MAX_IMAGE_WIDTH  = 4096             // 4096 pixels
	MAX_IMAGE_HEIGHT = 4096             // 4096 pixels
)

var ErrUnsupportedFormat = errors.New("metadata cleaning not supported for this format")

// Supported image MIME types
var supportedImageMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// Supported document MIME types
var supportedDocumentMIMEs = map[string]bool{
	"application/pdf":    true,
	"application/msword": true, // .doc
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true, // .docx
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

// Clean image metadata
//
// Skip unsupported formats.
// Return error if any
func CleanImageMetadata(data []byte) (cleanData []byte, format string, err error) {
	reader := bytes.NewReader(data)

	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, "", fmt.Errorf("could not decode image: %w", err)
	}

	isSupported := format == "jpeg" || format == "png" || format == "webp"
	if !isSupported {
		return nil, format, ErrUnsupportedFormat
	}

	var cleanBuf bytes.Buffer

	switch format {
	case "jpeg":
		if err := jpeg.Encode(&cleanBuf, img, nil); err != nil {
			return nil, "", fmt.Errorf("could not encode image: %w", err)
		}
	case "png":
		if err := png.Encode(&cleanBuf, img); err != nil {
			return nil, "", fmt.Errorf("could not encode image: %w", err)
		}
	case "webp":
		if err := webp.Encode(&cleanBuf, img, nil); err != nil {
			return nil, "", fmt.Errorf("could not encode image: %w", err)
		}
	}

	return cleanBuf.Bytes(), format, nil
}

// IsValidFile validates the file content against the expected file category.
func IsValidFile(data []byte, fileCategory model.FileCategory) (bool, error) {
	fileSize := len(data)
	switch fileCategory {
	case model.FileCategoryImage:
		if fileSize > MAX_IMAGE_SIZE {
			return false, fmt.Errorf("file size exceeds the maximum limit of %d bytes", MAX_IMAGE_SIZE)
		}
	case model.FileCategoryDocument:
		if fileSize > MAX_DOCS_SIZE {
			return false, fmt.Errorf("file size exceeds the maximum limit of %d bytes", MAX_DOCS_SIZE)
		}
	default:
		return false, fmt.Errorf("invalid file category: %s", fileCategory)
	}

	contentType := http.DetectContentType(data)

	if contentType == "application/octet-stream" || strings.HasPrefix(contentType, "text/") {
		return false, ErrUnsupportedFormat
	}

	switch fileCategory {
	case model.FileCategoryImage:
		if !supportedImageMIMEs[contentType] {
			return false, ErrUnsupportedFormat
		}

		reader := bytes.NewReader(data)
		config, _, err := image.DecodeConfig(reader)
		if err != nil {
			return false, fmt.Errorf("invalid image file: could not read dimensions or is corrupt: %w", err)
		}
		// Reject images whose width or height exceed configured maxima.
		if config.Width > MAX_IMAGE_WIDTH || config.Height > MAX_IMAGE_HEIGHT {
			return false, fmt.Errorf(
				"image dimensions exceed the maximum limit of %dx%d pixels",
				MAX_IMAGE_WIDTH,
				MAX_IMAGE_HEIGHT,
			)
		}

	case model.FileCategoryDocument:
		if !supportedDocumentMIMEs[contentType] {
			return false, ErrUnsupportedFormat
		}
	}

	return true, nil
}
