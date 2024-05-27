package models

import "time"

type Folder struct {
	DirName string
	CreatedAt time.Time
	UpdatedAt time.Time
}