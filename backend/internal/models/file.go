package models

import "gorm.io/gorm"

type FileUpdateBody struct {
	FileName   string `validate:"required" json:"file_name"`
	IsFavorite bool   `validate:"boolean" json:"is_favorite"`
	Restore    bool   `validate:"boolean" json:"is_restore"`
}

type FilePatchBody struct {
	FileName   string `json:"file_name"`
	FolderCode string `validate:"ascii" json:"folder_code"`
	IsFavorite bool   `validate:"boolean" json:"is_favorite"`
	Restore    bool   `validate:"boolean" json:"is_restore"`
}

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
	Thumbnail     *Thumbnail `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE;"`
}
