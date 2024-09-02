package services

import (
	"errors"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type ThumbnailService struct {
	DB *gorm.DB
	BucketClient *models.BucketClient
}

func (ts *ThumbnailService) SetDB(db *gorm.DB) {
	ts.DB = db
}

func (ts *ThumbnailService) SetBucketClient(bc *models.BucketClient) {
	ts.BucketClient = bc
}

func NewThumbnailService(db *gorm.DB, bc *models.BucketClient) *ThumbnailService {
	return &ThumbnailService{
		DB: db,
		BucketClient: bc,
	}
}

func (ts *ThumbnailService) DeleteThumbnail(thumbnail *models.Thumbnail) error {
	if err := ts.DB.Unscoped().Delete(thumbnail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "thumbnail not found",
				},
			}
		}

		return &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}	

	if err := ts.BucketClient.RemoveObject(thumbnail.FilePath, minio.RemoveObjectOptions{}); err != nil {
		return &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	return nil
}