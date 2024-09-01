package services

import (
	"log"
	"fmt"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"gorm.io/gorm"
)

type FileService struct {
	DB *gorm.DB
	BucketClient *models.BucketClient
}

func (fs *FileService) SetDB(db *gorm.DB) {
	fs.DB = db
}

func (fs *FileService) SetBucketClient(bc *models.BucketClient) {
	fs.BucketClient = bc
}

func NewFileService(db *gorm.DB) *FileService {
	// BucketClient is undefined because file service are initialized on main,
	// while BucketClient only available after JWTMiddleware
	return &FileService{
		DB: db,
	}
}

func (fs *FileService) ListFiles(userID uint, isTrashCan, isFavorite bool) ([]models.File, error) {
	if isTrashCan && isFavorite {
		return nil, &apperr.InvalidParamError{
			BaseError: &apperr.BaseError{
				Message: "cannot fetch trash can and favorite at the same time",
				Err: fmt.Errorf("cannot fetch trash can and favorite at the same time"),
			},
		}
	}

	if isFavorite {
		var favoriteFiles []models.File
		if err := fs.DB.Where("user_id = ? AND is_favorite = ?", userID, true).Find(&favoriteFiles).Error; err != nil {
			log.Println(err.Error())
			return nil, &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err: err,
				},
			}
		}

		return favoriteFiles, nil
	}

	// Trash can
	var trashedFiles []models.File
	if err := fs.DB.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", userID).Find(&trashedFiles).Error; err != nil {
		log.Println(err.Error())
		return nil, &apperr.InvalidParamError{
			BaseError: &apperr.BaseError{
				Message: "cannot fetch trash can and favorite at the same time",
				Err: fmt.Errorf("cannot fetch trash can and favorite at the same time"),
			},
		}
	}

	return trashedFiles, nil
}