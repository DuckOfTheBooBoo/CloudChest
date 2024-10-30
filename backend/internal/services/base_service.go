package services

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"gorm.io/gorm"
)

type Service interface {
	SetDB(db *gorm.DB)
	SetBucketClient(bc *models.BucketClient)
}


