package repository

import "ku-work/backend/model"

type FileRepository interface {
	Save(file *model.File) error
	Delete(file *model.File) error
	Get(file *model.File) (*model.File, error)
	GetByID(fileID string) (*model.File, error)
}
