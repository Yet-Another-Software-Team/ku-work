package model

import (
	"time"
)

type FileType string

const (
	// Image file types
	FileTypeJPG  FileType = "jpg"
	FileTypeJPEG FileType = "jpeg"
	FileTypePNG  FileType = "png"
	FileTypeWEBP FileType = "webp"

	// Document file types
	FileTypePDF  FileType = "pdf"
	FileTypeDOC  FileType = "doc"
	FileTypeDOCX FileType = "docx"
)

type FileCategory string

const (
	FileCategoryImage    FileCategory = "image"
	FileCategoryDocument FileCategory = "document"
)

// File represents a file stored in the system
type File struct {
	ID        string       `gorm:"type:uuid;primarykey;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	UserID    string       `gorm:"type:uuid;not null" json:"userId"`
	Category  FileCategory `gorm:"not null" json:"category"`
}

// Note: file-level storage operations are handled via storage-provider hooks.
// The FileType enum/consts exist for callers that need to represent or detect
// an on-disk/content file type. The File struct intentionally does not force
// storage-specific behavior; actual deletion of stored objects is done by
// higher-level model hooks (Company, Student, JobApplication) via the
// registered storage deletion hook.
