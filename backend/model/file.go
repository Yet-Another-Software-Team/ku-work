package model

import (
	"os"
	"path/filepath"
	"time"

	"gorm.io/gorm"
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
	FileType  FileType     `gorm:"not null" json:"fileType"`
	Category  FileCategory `gorm:"not null" json:"category"`
}

func (file *File) AfterDelete(tx *gorm.DB) (err error) {
	if file.ID == "" {
		return nil
	}
	return os.Remove(filepath.Join("./files", file.ID))
}
