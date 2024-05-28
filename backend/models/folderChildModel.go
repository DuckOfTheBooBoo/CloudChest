package models

import "time"

type FolderChild struct {	
	UserID uint   `gorm:"not null;primaryKey"`
	Parent string `gorm:"type:varchar(255);not null;primaryKey"`
	Child  string `gorm:"type:varchar(255);not null;primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
