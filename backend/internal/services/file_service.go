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

// NewFileService creates a new FileService with the given database connection.
//
// Note that the BucketClient field of the returned FileService is nil, because
// it is only available after the JWTMiddleware has been called. This is
// necessary because the JWTMiddleware is responsible for setting the
// BucketClient on the Gin context.
func NewFileService(db *gorm.DB) *FileService {
	// BucketClient is undefined because file service are initialized on main,
	// while BucketClient only available after JWTMiddleware
	return &FileService{
		DB: db,
	}
}

// ListFiles lists all files of a user, with the given params.
//
// If isTrashCan is true, it will list all files in the user's trash can.
// If isFavorite is true, it will list all files favorited by the user.
//
// If both isTrashCan and isFavorite are true, it returns an error.
//
// Note that the returned files are not sorted in any particular order.
//
// If an internal error occurs, it will return a ServerError.
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

// DeleteFileTemp soft deletes a file by setting its deleted_at field to the current time.
// It returns an error if the file does not exist, or if there was an internal server error.
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

// DeleteFilePermanent permanently deletes a file by deleting the file itself and its
// associated objects such as thumbnails and HLS files. It returns an error if the file
// does not exist, or if there was an internal server error.
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

// EmptyTrashCan deletes all files in a user's trash can.
// It returns an error if there was an internal server error.
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

// PatchFile updates a file by given file id and user id.
// 
// This function also supports restoring a file from trash can by setting Restore to true.
// 
// If the file is not found, it returns a NotFoundError.
// If other errors occur, it returns a ServerError.
func (fs *FileService) PatchFile(userID, fileID uint, patchBody models.FilePatchBody) (*models.File, error) {
	// Find file
	var file models.File
	query := fs.DB.Where("id = ? AND user_id = ?", fileID, userID)

	if patchBody.Restore {
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

	if file.FileName != patchBody.FileName {
		file.FileName = patchBody.FileName
	}

	if file.IsFavorite != patchBody.IsFavorite {
		file.IsFavorite = patchBody.IsFavorite
	}

	if patchBody.Restore {
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