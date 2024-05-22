package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Username    string `gorm:"type:varchar(50)"`
	Email       string `gorm:"unique;type:varchar(255)"`
	Password    string `json:"-" gorm:"min:6;type:varchar(64)"`
	MinioBucket string `gorm:"type:varchar(50)"`
}
