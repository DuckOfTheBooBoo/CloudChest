package models

import "gorm.io/gorm"

type Thumbnail struct {
	gorm.Model
	FileID   uint   `gorm:"not null"`
	FilePath string `gorm:"type:varchar(255);not null"`
}
