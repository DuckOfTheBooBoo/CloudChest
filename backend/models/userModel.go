package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	FirstName   string `gorm:"type:varchar(50)"`
	LastName    string `gorm:"type:varchar(50)"`
	Email       string `gorm:"unique;type:varchar(255)"`
	Password    string `json:"-" gorm:"min:6;type:varchar(64)"`
	MinioBucket string `gorm:"type:varchar(50)"`
	Files       []*File
	FolderChild []*FolderChild
}
