package services

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/minio/minio-go/v7"
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

func (fs *FileService) DeleteFileTemp(userID, fileID uint) error {
	if err := fs.DB.Where("user_id = ? AND id = ?", userID, fileID).Delete(&models.File{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "file not found",
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

	return nil
}

func (fs *FileService) DeleteFilePermanent(userID, fileID uint) error {
	var file models.File
	if err := fs.DB.Unscoped().Where("user_id = ? AND id = ?", userID, fileID).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "file not found",
					Err: err,
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

	err := fs.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Delete(file).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &apperr.NotFoundError{
					BaseError: &apperr.BaseError{
						Message: "file not found",
						Err: err,
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
	
		if strings.HasPrefix(file.FileType, "image/") || strings.HasPrefix(file.FileType, "video/") {
			thumbnailService := NewThumbnailService(fs.DB, fs.BucketClient)
			if err := thumbnailService.DeleteThumbnail(file.Thumbnail); err != nil {
				return err
			}
			
			if strings.HasPrefix(file.FileType, "video/") && file.IsPreviewable {
				hlsService := NewHLSService(fs.DB, fs.BucketClient)
				hlsService.DeleteHLSFiles(&file)
			}
		}
	
		if err := fs.BucketClient.RemoveObject(file.FileCode, minio.RemoveObjectOptions{}); err != nil {
			return &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err: err,
				},
			}
		}
	
		return nil
	})

	return err
}

func (fs *FileService) EmptyTrashCan(userID uint) error {
	var deletedFiles []models.File
	if err := fs.DB.Unscoped().Preload("Thumbnail").Where("user_id = ? AND deleted_at IS NOT NULL", userID).Find(&deletedFiles).Error; err != nil {
		return err
	}

	for _, file := range deletedFiles {
		if err := fs.DeleteFilePermanent(userID, file.ID); err != nil {
			return &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err: err,
				},
			}
		}
	}
	return nil
}

// UpdateFile updates a file by given file id and user id.
//
// This function also supports restoring a file from trash can by setting Restore to true.
//
// If the file is not found, it returns a NotFoundError.
// If other errors occur, it returns a ServerError.
func (fs *FileService) UpdateFile(userID, fileID uint, updateBody models.FileUpdateBody) (*models.File, error) {
	// Find file
	var file models.File
	query := fs.DB.Where("id = ? AND user_id = ?", fileID, userID)

	if updateBody.Restore {
		query = query.Unscoped()
	}
	
	if err := query.First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "File not found",
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

	file.FileName = updateBody.FileName
	file.IsFavorite = updateBody.IsFavorite

	if updateBody.Restore {
		if err := fs.DB.Unscoped().Model(&file).Update("deleted_at", nil).Error; err != nil {
			return nil, &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err: err,
				},
			}
		}
	}

	if err := fs.DB.Save(&file).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	return &file, nil
}