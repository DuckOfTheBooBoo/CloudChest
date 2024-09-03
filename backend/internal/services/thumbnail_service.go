package services

import (
	"errors"
	"strings"

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

func (ts *ThumbnailService) GetThumbnail(fileCode string, userID uint) (*minio.Object, error) {
	var file models.File
	if err := ts.DB.Model(&models.File{}).Where("file_code = ? AND user_id = ?", fileCode, userID).Preload("Thumbnail").First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "file not found",
					Err: err,
				},
			}
		}

		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	if file.Thumbnail == nil {
		if strings.HasPrefix(file.FileType, "image/") {
			return nil, &apperr.ResourceNotReadyError{
				BaseError: &apperr.BaseError{
					Message: "file's thumbnail is being processed",
				},
			}
		} else {
			return nil, &apperr.InvalidParamError{
				BaseError: &apperr.BaseError{
					Message: "file is not an image or a video",
				},
			}
		}
	}

	// Close at handler
	thumbnail, err := ts.BucketClient.GetObject(file.Thumbnail.FilePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}
	
	return thumbnail, nil
}