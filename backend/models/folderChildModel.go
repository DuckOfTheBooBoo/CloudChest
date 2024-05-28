package models

type FolderChild struct {
	UserID uint   `gorm:"not null;primaryKey"`
	Parent string `gorm:"type:varchar(255);not null;primaryKey"`
	Child  string `gorm:"type:varchar(255);not null;primaryKey"`
}
