package models

import "time"

type Token struct {
	Token  string `gorm:"type:varchar(255)"`
	ExpirationDate time.Time
}
