package model

import (
	"gorm.io/datatypes"
	"time"
)

type Student struct {
	UserID            string `gorm:"type:uuid;primarykey"`
	User              User   `gorm:"foreignKey:UserID"`
	Approved          bool
	CreatedAt         time.Time
	FullName          string
	Phone             string
	Photo             string
	BirthDate         datatypes.Date
	AboutMe           string
	GitHub            string
	LinkedIn          string
	StudentID         string
	Major             string
	StudentStatus     string
	StudentStatusFile string
}
