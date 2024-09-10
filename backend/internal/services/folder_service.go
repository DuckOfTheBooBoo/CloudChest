package services

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/gofrs/uuid/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
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

const (
	MAX_PREVIEWABLE_VIDEO_SIZE = 150 * 1000 * 1000
)

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

// PatchFolder updates a folder. If folderUpdateBody.Restore is true, it will restore a soft-deleted folder.
func (fs *FolderService) PatchFolder(userID uint, folderCode string, folderUpdateBody models.FolderUpdateBody) (*models.Folder, error) {
	var folder models.Folder

	query := fs.DB.Where("code = ? AND user_id = ?", folderCode, userID)

	if folderUpdateBody.Restore {
		query = query.Unscoped()
	}

	if err := query.First(&folder).Error; err != nil {
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
				Message: "Failed to patch folder",
				Err: err,
			},
		}
	}

	if folderUpdateBody.FolderName != "" {
		folder.Name = folderUpdateBody.FolderName
	}

	if folder.IsFavorite != folderUpdateBody.IsFavorite {
		folder.IsFavorite = folderUpdateBody.IsFavorite
	}

	if folderUpdateBody.ParentFolderCode != "" {
		var parentFolder models.Folder
		if err := fs.DB.Where("code = ? AND user_id = ?", folderUpdateBody.ParentFolderCode, userID).First(&parentFolder).Error; err != nil {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "Folder not found",
					Err: err,
				},
			}
		}

		folder.ParentID = &parentFolder.ID
	}

	if folderUpdateBody.Restore {
		if err := fs.DB.Unscoped().Model(&folder).Update("deleted_at", nil).Error; err != nil {
			return nil, &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Failed to restore folder " + folder.Name,
					Err: err,
				},
			}
		}
	}

	if err := fs.DB.Save(&folder).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to update folder " + folder.Name,
				Err: err,
			},
		}
	}

	return &folder, nil
}

// FetchFolderFiles fetches all files in the given folder.
//
// If the folder is not found, it returns a NotFoundError.
// If other errors occur, it returns a ServerError.
func (fs *FolderService) FetchFolderFiles(userID uint, folderCode string) ([]*models.File, error) {
	if folderCode == "root" {
		folderCode = ""
	}
	
	var parentFolder models.Folder
	if err := fs.DB.Where("user_id = ? AND code = ?", userID, folderCode).Preload("Files").First(&parentFolder).Error; err != nil {
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
				Message: "Failed to fetch folder files",
				Err: err,
			},
		}
	}

	return parentFolder.Files, nil
}

func (fs *FolderService) UploadFile(userID uint, folderCode string, file *multipart.FileHeader) (*models.File, []byte, error) {
	query := fs.DB.Where("user_id = ? AND code = ?", userID, folderCode)

	if folderCode == "root" {
		query = fs.DB.Where("user_id = ? AND (code IS NULL OR code = '')", userID)
	}

	var parentFolder models.Folder
	if err := query.First(&parentFolder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "Folder not found",
					Err: err,
				},
			}
		}

		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to upload file",
				Err: err,
			},
		}
	}


	fileCode, err := uuid.NewV4()
	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to generate file code",
				Err: err,
			},
		}
	}

	// UPLOAD FILE RECORD TO RDBMS
	// Create new File record in rbdms
	newFile := models.File{
		UserID:     userID,
		FolderID:   parentFolder.ID,
		FileName:   file.Filename,
		FileCode:   fileCode.String(),
		FileSize:   uint(file.Size),
		FileType:   file.Header.Get("Content-Type"),
		IsFavorite: false,
	}

	if strings.HasPrefix(newFile.FileType, "image/") {
		newFile.IsPreviewable = true
	}

	// UPLOAD FILE TO MINIO
	// Read the file
	uploadedFile, err := file.Open()
	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to open file",
				Err: err,
			},
		}
	}

	// Shit happens, uploadedFile have a body length of 0? It went well until I pass it into minioClient for uploading to the bucket. The current workaround is to convert it into []byte regardless its mime type
	// Prepare []byte of the uploded file in case its a media (image or video) file. Else let it collected by garbage collector
	var uploadedFileBytes []byte
	uploadedFileBytes, err = io.ReadAll(uploadedFile)
	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to read file",
				Err: err,
			},
		}
	}
	uploadedFile.Close()

	filePath := "/" + fileCode.String()

	// Use transaction, PutObject to minio could lead to an error. If it does, we can't let any changes happen in the database
	err = fs.DB.Transaction(func(tx *gorm.DB) error {
		// Upload the file to minio first
		_, err = fs.BucketClient.PutObject(filePath, bytes.NewReader(uploadedFileBytes), file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
		if err != nil {
			return fmt.Errorf("error while uploading file to MinIO: %v", err)
		}

		if err := tx.Create(&newFile).Error; err != nil {
			if err := fs.BucketClient.RemoveObject(filePath, minio.RemoveObjectOptions{}); err != nil {
				return fmt.Errorf("error while undoing MinIO file uploading: %v", err)
			}
			return fmt.Errorf("error while creating file in database: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to upload file",
				Err: err,
			},
		}
	}

	return &newFile, uploadedFileBytes, nil
}

func (fs *FolderService) PostUploadProcess(file *models.File, uploadedFileBytes []byte) {
	ctx := context.Background()
	go func() {
		jobCtx, cancel := context.WithCancel(ctx)
		defer cancel()
		var wg sync.WaitGroup
		tempThumbFilePathChan := make(chan string, 1)
		tempHLSbFilePathChan := make(chan string)

		// Write file to temp dir
		wg.Add(1)
		go func() {
			defer wg.Done()

			tempPath := fmt.Sprintf("/tmp/%s-file", file.FileCode)
			tempFile, err := os.Create(tempPath)
			if err != nil {
				log.Printf("Error while creating temp file: %v", err)
				cancel()
			}

			_, err = tempFile.Write(uploadedFileBytes)
			if err != nil {
				log.Printf("Error while writing temp file: %v", err)
				cancel()
			}
		
			tempThumbFilePathChan <- tempPath
			tempHLSbFilePathChan <- tempPath
		}()

		// Process thumbnail
		wg.Add(1)
		go func(ctx context.Context) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				log.Println("Thumbnail generation cancelled due to error in writing temp file.")
				return

			default:
				filePath := <-tempThumbFilePathChan
				thumbnailService := NewThumbnailService(fs.DB, fs.BucketClient)
				thumbnailService.GenerateThumbnail(filePath, file)
			}

		}(jobCtx)

		// Process HLS file (video only)
		if strings.HasPrefix(file.FileType, "video/") && file.FileSize <= MAX_PREVIEWABLE_VIDEO_SIZE {
			log.Printf("Processing %s for HLS", file.FileName)
			wg.Add(1)
			go func(ctx context.Context) {
				defer wg.Done()
				select {
				case <-ctx.Done():
					log.Println("HLS process cancelled due to error in writing temp file.")
					return

				default:
					filePath := <-tempHLSbFilePathChan
					hlsService := NewHLSService(fs.DB, fs.BucketClient)
					hlsService.ProcessHLS(filePath, file)
				}
			}(jobCtx)
		}

		// Remove temp file
		wg.Wait()
		filePath := <-tempThumbFilePathChan
		os.Remove(filePath)
		log.Println("Removed temp file: " + filePath)
	}()
}

func (fs *FolderService) CreateFolder(folderName, parentFolderCode string, userID uint) (*models.Folder, error) {
	newFolderCode, err := gonanoid.New()
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to generate folder code",
				Err: err,
			},
		}
	}

	query := fs.DB.Where("user_id = ? AND (code IS NULL OR code = '')", userID)

	if parentFolderCode != "root" {
		query = fs.DB.Where("user_id = ? AND code = ?", userID, parentFolderCode)
	}

	// Fetch parent folder
	var parentFolder models.Folder
	// Query by parent folder code
	if query.First(&parentFolder).Error != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "Parent folder not found",
					Err: err,
				},
			}
		}

		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to fetch parent folder",
				Err: err,
			},
		}
	}

	newFolder := models.Folder{
		UserID:   userID,
		ParentID: &parentFolder.ID,
		Name:     folderName,
		Code:     newFolderCode,
	}

	if err := fs.DB.Create(&newFolder).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to create folder",
				Err: err,
			},
		}
	}

	parentFolder.HasChild = true
	if err := fs.DB.Save(&parentFolder).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to update parent folder",
				Err: err,
			},
		}
	}

	return &newFolder, nil
}