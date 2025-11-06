package gormrepo

import (
	"errors"

	"ku-work/backend/model"
	repo "ku-work/backend/repository"

	"gorm.io/gorm"
)

var _ repo.FileRepository = (*GormFileRepository)(nil)

type GormFileRepository struct {
	db *gorm.DB
}

func NewGormFileRepository(db *gorm.DB) repo.FileRepository {
	return &GormFileRepository{db: db}
}

func (r *GormFileRepository) Save(file *model.File) error {
	if r == nil || r.db == nil {
		return errors.New("file repository not initialized")
	}
	if file == nil {
		return gorm.ErrInvalidData
	}
	return r.db.Create(file).Error
}

func (r *GormFileRepository) Delete(file *model.File) error {
	if r == nil || r.db == nil {
		return errors.New("file repository not initialized")
	}
	if file == nil {
		return gorm.ErrInvalidData
	}
	if file.ID != "" {
		return r.db.Where("id = ?", file.ID).Delete(&model.File{}).Error
	}
	return r.db.Delete(file).Error
}

func (r *GormFileRepository) Get(file *model.File) (*model.File, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("file repository not initialized")
	}
	if file == nil {
		return nil, gorm.ErrInvalidData
	}

	var out model.File
	q := r.db.Model(&model.File{})
	var err error
	if file.ID != "" {
		err = q.Where("id = ?", file.ID).First(&out).Error
	} else {
		err = q.Where(file).First(&out).Error
	}
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *GormFileRepository) GetByID(fileID string) (*model.File, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("file repository not initialized")
	}
	var out model.File
	if err := r.db.Where("id = ?", fileID).First(&out).Error; err != nil {
		return nil, err
	}
	return &out, nil
}
