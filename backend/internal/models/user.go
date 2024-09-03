package models

import "gorm.io/gorm"

type UserBody struct {
	FirstName string `json:"first_name" validate:"required,ascii"`
	LastName  string `json:"last_name" validate:"required,ascii"`
	Email     string `validate:"required,email"`
	Password  string `validate:"required,min=6"`
}

type User struct {
	gorm.Model
	FirstName          string `gorm:"type:varchar(50)"`
	LastName           string `gorm:"type:varchar(50)"`
	Email              string `gorm:"unique;type:varchar(255)"`
	Password           string `json:"-" gorm:"min:6;type:varchar(64)"`
	MinioBucket        string `json:"-"`
	MinioServiceBucket string `json:"-"`
	Folders            []*Folder
	Files              []*File
}
