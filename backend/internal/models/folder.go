package models

import (
	"gorm.io/gorm"
)

type Folder struct {
	gorm.Model
	UserID       uint `gorm:"not null"`
	ParentID     *uint
	Name         string    `gorm:"type:varchar(255);not null"`
	Code         string    `gorm:"type:varchar(100)"`
	HasChild     bool      `gorm:"default:0"`
	IsFavorite   bool      `gorm:"default:0"`
	ChildFolders []*Folder `gorm:"foreignKey:ParentID"`
	Files        []*File   `gorm:"foreignKey:FolderID"`
}

type FolderHierarchy struct {
	Name string `json:"name"`
	Code string `json:"code"`
}