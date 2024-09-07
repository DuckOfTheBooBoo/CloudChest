package services

import (
	"errors"
	"slices"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"gorm.io/gorm"
)

type FolderService struct {
	DB *gorm.DB
	BucketClient *models.BucketClient
}

func NewFolderService(db *gorm.DB) *FolderService {
	return &FolderService{
		DB: db,
	}
}

func (fs *FolderService) SetDB(db *gorm.DB) {
	fs.DB = db
}

func (fs *FolderService) SetBucketClient(bc *models.BucketClient) {
	// BucketClient is nil because folder service are initialized on main,
	// while BucketClient only available after JWTMiddleware
	fs.BucketClient = bc
}

type FolderResponse struct {
	Folders []*models.Folder `json:"folders"`
	Hierarchies []models.FolderHierarchy `json:"hierarchies"`
}

// ListFolders lists all folders of a user, with the given params.
//
// If folderCode is "root", it will list all top-level folders.
// If folderCode is not "root", it will list all subfolders of the folder with the given folderCode.
//
// The returned FolderResponse contains a list of folders, and their hierarchy.
//
// If the folder is not found, it returns a NotFoundError.
// If other errors occur, it returns a ServerError.
func (fs *FolderService) ListFolders(userID uint, folderCode string) (*FolderResponse, error) {

	var parentFolder models.Folder
	query := fs.DB.Where("user_id = ? AND (code IS NULL OR code = '')", userID)
	if folderCode != "root" {
		query = fs.DB.Where("user_id = ? AND code = ?", userID, folderCode)
	}

	if err := query.Preload("ChildFolders").Find(&parentFolder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "Folder not found",
					Err: err,
				},
			}
		}

		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to list folders",
				Err: err,
			},
		}
	}
		
	// Generate hierarchy
	var currentParent *models.Folder = &parentFolder
	var hierarchies []models.FolderHierarchy
	for currentParent.ParentID != nil {
		var parent models.Folder
		if err := fs.DB.Where("id = ?", *currentParent.ParentID).Find(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, &apperr.NotFoundError{
					BaseError: &apperr.BaseError{
						Message: "Folder not found",
						Err: err,
					},
				}

			}
			return nil, &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Failed to list folders",
					Err: err,
				},
			}
		}
		folderHierarchy := models.FolderHierarchy{
			Name: parent.Name,
			Code: parent.Code,
		}
		hierarchies = append(hierarchies, folderHierarchy)
		currentParent = &parent
	}

	slices.Reverse(hierarchies)
	hierarchies = append(hierarchies, models.FolderHierarchy{
		Name: parentFolder.Name,
		Code: parentFolder.Code,
	})

	return &FolderResponse{
		Folders: parentFolder.ChildFolders,
		Hierarchies: hierarchies,
	}, nil
}
