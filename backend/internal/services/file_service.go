package services

import (
	"errors"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type FileService struct {
	DB           *gorm.DB
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

// ListFavoriteFiles lists all files of a user that are marked as favorite.
//
// If an internal server error occurs, it returns a ServerError.
// Otherwise, it returns a list of favorite files of the user.
func (fs *FileService) ListFavoriteFiles(userID uint) ([]models.File, error) {
	var favoriteFiles []models.File
	if err := fs.DB.Where("user_id = ? AND is_favorite = ?", userID, true).Find(&favoriteFiles).Error; err != nil {
		log.Println(err.Error())
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err:     err,
			},
		}
	}

	return favoriteFiles, nil
}

// ListTrashCanFiles lists all files of a user that are trashed.
//
// If an internal server error occurs, it returns a ServerError.
// Otherwise, it returns a list of trashed files of the user.
func (fs *FileService) ListTrashCanFiles(userID uint) ([]models.File, error) {
	var trashedFiles []models.File
	if err := fs.DB.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", userID).Find(&trashedFiles).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to fetch trashed files",
				Err:     err,
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
				Err:     err,
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
					Err:     err,
				},
			}
		}

		return &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err:     err,
			},
		}
	}

	err := fs.DB.Transaction(func(tx *gorm.DB) error {
		if strings.HasPrefix(file.FileType, "image/") || strings.HasPrefix(file.FileType, "video/") {
			if file.Thumbnail != nil {
				thumbnailService := NewThumbnailService(fs.DB, fs.BucketClient)
				if err := thumbnailService.DeleteThumbnail(file.Thumbnail); err != nil {
					if !errors.Is(err, &apperr.NotFoundError{}) {
						return err
					}
				}
			}

			if strings.HasPrefix(file.FileType, "video/") && file.IsPreviewable {
				hlsService := NewHLSService(fs.DB, fs.BucketClient)
				hlsService.DeleteHLSFiles(&file)
			}
		}

		if err := tx.Unscoped().Delete(file).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return &apperr.NotFoundError{
					BaseError: &apperr.BaseError{
						Message: "file not found",
						Err:     err,
					},
				}
			}

			return &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err:     err,
				},
			}
		}

		if err := fs.BucketClient.RemoveObject(file.FileCode, minio.RemoveObjectOptions{}); err != nil {
			return &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err:     err,
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
					Err:     err,
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
	query := fs.DB.Where("id = ? AND user_id = ?", fileID, userID).Preload("Folder")

	if updateBody.Restore {
		query = query.Unscoped()
	}

	if err := query.First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "File not found",
					Err:     err,
				},
			}
		}

		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err:     err,
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
					Err:     err,
				},
			}
		}

		if file.Folder.DeletedAt.Valid {
			folderService := NewFolderService(fs.DB)
			if err := folderService.RecursivelyRestoreFoldersUpwards(file.Folder); err != nil {
				return nil, &apperr.ServerError{
					BaseError: &apperr.BaseError{
						Message: "failed to restore file's parent folder",
						Err:     err,
					},
				}
			}
		}
	}

	if err := fs.DB.Save(&file).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err:     err,
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
	query := fs.DB.Where("id = ? AND user_id = ?", fileID, userID).Preload("Folder")

	if patchBody.Restore {
		query = query.Unscoped()
	}

	if err := query.First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "File not found",
					Err:     err,
				},
			}
		}

		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err:     err,
			},
		}
	}

	if file.FileName != patchBody.FileName && patchBody.FileName != "" {
		file.FileName = patchBody.FileName
	} else if file.IsFavorite != patchBody.IsFavorite && patchBody.IsFavorite {
		// Why else if? Because patch request can only set one field at a time.
		// if a user requests a file rename, they will not set a value to the rest of the fields
		// in this case, patchBody.IsFavorite. Therefore, its default value is false.
		// This might lead to a file being set as "not favorite" even though user only requests a rename
		file.IsFavorite = patchBody.IsFavorite
	}

	if patchBody.Restore {
		if err := fs.DB.Unscoped().Model(&file).Update("deleted_at", nil).Error; err != nil {
			return nil, &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err:     err,
				},
			}
		}

		if file.Folder.DeletedAt.Valid {
			folderService := NewFolderService(fs.DB)
			if err := folderService.RecursivelyRestoreFoldersUpwards(file.Folder); err != nil {
				return nil, &apperr.ServerError{
					BaseError: &apperr.BaseError{
						Message: "failed to restore file's parent folder",
						Err:     err,
					},
				}
			}
		}
	}

	if err := fs.DB.Save(&file).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err:     err,
			},
		}
	}

	return &file, nil
}

func (fs *FileService) GetPresignedURL(userID uint, fileCode string) (*url.URL, error) {
	var file models.File
	if err := fs.DB.Where("file_code = ? AND user_id = ?", fileCode, userID).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "File not found",
					Err:     err,
				},
			}
		}

		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to fetch file's information",
				Err:     err,
			},
		}
	}

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\""+file.FileName+"\"")

	charsetParam := ""
	if strings.HasPrefix(file.FileType, "text/") {
		charsetParam = "; charset=utf-8"
	}

	reqParams.Set("response-content-type", file.FileType+charsetParam)

	presignedURL, err := fs.BucketClient.PresignedGetObject(file.FileCode, time.Second*24*60*60, reqParams)
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to get presigned URL",
				Err:     err,
			},
		}
	}

	return presignedURL, nil
}
