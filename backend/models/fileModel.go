package models

import "gorm.io/gorm"

type File struct {
	gorm.Model
	UserID        uint       `gorm:"not null"`
	FolderID      uint       `gorm:"not null"`
	FileName      string     `gorm:"type:varchar(255);not null"`
	FileCode      string     `gorm:"type:char(36);not null"`
	FileSize      uint       `gorm:"not null"`
	FileType      string     `gorm:"type:varchar(100);not null"`
	IsFavorite    bool       `gorm:"not null;default:0"`
	IsPreviewable bool       `gorm:"not null;default:0"`
	Folder        *Folder    `gorm:"foreignKey:FolderID"`
	Thumbnail     *Thumbnail `gorm:"foreignKey:FileID"`
}
