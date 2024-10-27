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
	"gorm.io/gorm/clause"
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
func (fs *FolderService) ListFolders(userID uint, folderCode string) (*models.FolderResponse, error) {

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

	return &models.FolderResponse{
		Folders: parentFolder.ChildFolders,
		Hierarchies: hierarchies,
	}, nil
}

func (fs *FolderService) ListFavoriteFolders(userID uint) (*models.FolderResponse, error) {
	var favoriteFolders []*models.Folder
	if err := fs.DB.Where("user_id = ? AND is_favorite = ?", userID, true).Find(&favoriteFolders).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to list favorite folders",
				Err: err,
			},
		}
	}

	return &models.FolderResponse{
		Folders: favoriteFolders,
	}, nil
}

func (fs *FolderService) ListTrashFolders(userID uint) (*models.FolderResponse, error) {
	var trashFolders []*models.Folder
	if err := fs.DB.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", userID).Find(&trashFolders).Error; err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to list trash folders",
				Err: err,
			},
		}
	}

	return &models.FolderResponse{
		Folders: trashFolders,
	}, nil
}

// PatchFolder updates a folder. If folderUpdateBody.Restore is true, it will restore a soft-deleted folder.
func (fs *FolderService) PatchFolder(userID uint, folderCode string, folderUpdateBody models.FolderUpdateBody) (*models.Folder, error) {
	var folder models.Folder

	query := fs.DB.Where("code = ? AND user_id = ?", folderCode, userID).Preload("ParentFolder")

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
		folder.ParentFolder = &parentFolder
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

		if folder.ParentFolder.DeletedAt.Valid {
			if err := fs.RecursivelyRestoreFoldersUpwards(folder.ParentFolder); err != nil {
				return nil, &apperr.ServerError{
					BaseError: &apperr.BaseError{
						Message: "Failed to restore folder " + folder.Name,
						Err: err,
					},
				}
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

// DeleteFolderTemp deletes a folder temporarily. If the folder is not found, it returns a NotFoundError.
// If other errors occur, it returns a ServerError.
func (fs *FolderService) DeleteFolderTemp(folderCode string, userID uint) error {
	targetFolder := models.Folder{
		UserID: userID,
		Code:   folderCode,
	}

	if err := fs.DB.Where("code = ? AND user_id = ?", folderCode, userID).Delete(&targetFolder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "Folder not found",
					Err: err,
				},
			}
		}

		return &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to delete folder",
				Err: err,
			},
		}
	}

	return nil
}


type DeletedFilesAndFoldersList struct {
	DeletedFiles []string `json:"deleted_files"`
	DeletedFolders []string `json:"deleted_folders"`
}

// DeleteFolderPermanent deletes a folder and all its contents permanently. If the folder is not found, it returns a NotFoundError.
// If other errors occur, it returns a ServerError.
func (fs *FolderService) DeleteFolderPermanent(folderCode string, userID uint) (*DeletedFilesAndFoldersList, error) {
	var targetFolder models.Folder
	if err := fs.DB.Unscoped().Where("code = ? AND user_id = ?", folderCode, userID).First(&targetFolder).Error; err != nil {
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
				Message: "Failed to delete folder",
				Err: err,
			},
		}
	}

	if err := fs.loadFolders(&targetFolder); err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to load folder",
				Err: err,
			},
		}
	}

	var deletedObjects DeletedFilesAndFoldersList = DeletedFilesAndFoldersList{
		DeletedFiles: []string{},
		DeletedFolders: []string{},
	}

	if err := fs.processFolder(&deletedObjects, &targetFolder); err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to delete folder",
				Err: err,
			},
		}
	}

	return &deletedObjects, nil
}

// Only used when we're trying to restore a deleted (temp) folder but its parent or grandparent or great-grandparent or whatever that is, is deleted (temp)
func (fs *FolderService) RecursivelyRestoreFoldersUpwards(parentFolder *models.Folder) error {
	// Restore the parent parentFolder if it was soft deleted
	if parentFolder.DeletedAt.Valid {
		log.Printf("Restoring folder %s (%s)\n", parentFolder.Name, parentFolder.Code)
		if err := fs.DB.Unscoped().Model(parentFolder).Update("deleted_at", nil).Error; err != nil {
			return err
		}
	}

	// Fetch the parent parentFolder
	if err := fs.DB.Unscoped().Model(&models.Folder{}).Where("id = ?", parentFolder.ParentID).First(&parentFolder.ParentFolder).Error; err != nil {
		return err
	}
	
	if parentFolder.ParentFolder.Name == "/" {
		return nil
	}
	
	// Recursively restore the parent's parent
	return fs.RecursivelyRestoreFoldersUpwards(parentFolder.ParentFolder)
}


// loadFolders preloads the immediate child folders of the given folder, and then
// recursively loads all the children of each child folder. It returns an error if
// any of the loads fail.
func (fs *FolderService) loadFolders(folder *models.Folder) error {
	// Preload immediate child folders
	if err := fs.DB.Unscoped().Preload(clause.Associations).Find(&folder).Error; err != nil {
		return err
	}

	// Recursively load children of each child folder
	for i := range folder.ChildFolders {
		if err := fs.loadFolders(folder.ChildFolders[i]); err != nil {
			return err
		}
	}

	return nil
}

// processFolder recursively deletes all files and child folders of the given folder.
// This is called by DeleteFolderPermanent.
func (fs *FolderService) processFolder(deletedObjects *DeletedFilesAndFoldersList, folder *models.Folder) error {
	bc := fs.BucketClient
	for _, child := range(folder.ChildFolders) {
		if err := fs.processFolder(deletedObjects, child); err != nil {
			return err
		}
	}

	var toBeDeletedFiles []*models.File
	var filesThumbnail []*models.Thumbnail
	for _, file := range(folder.Files) {
		log.Printf("Deleting file %s (%s)\n", file.FileName, file.FileCode)
		toBeDeletedFiles = append(toBeDeletedFiles, file)
		if file.Thumbnail != nil {
			filesThumbnail = append(filesThumbnail, file.Thumbnail)
		}
	}

	// Delete thumbnails from DB
	if len(filesThumbnail) > 0 {
		if err := fs.DB.Unscoped().Delete(&filesThumbnail).Error; err != nil {
			return err
		}
	
		thumbObjCh := make(chan minio.ObjectInfo)
		go func() {
			defer close(thumbObjCh)
			for _, thumb := range filesThumbnail {
				if len(toBeDeletedFiles) > 0 {
					obj := minio.ObjectInfo{
						Key: thumb.FilePath,
					}
					thumbObjCh <- obj
				}
			}
		}()
	
		for err := range bc.Client.RemoveObjects(bc.Context, bc.ServiceBucket, thumbObjCh, minio.RemoveObjectsOptions{}) {
			if err.Err != nil {
				return err.Err
			}
		}
	}

	// Delete files from DB
	if len(toBeDeletedFiles) > 0 {
		for deletedFile := range toBeDeletedFiles {
			deletedObjects.DeletedFiles = append(deletedObjects.DeletedFiles, toBeDeletedFiles[deletedFile].FileCode)
		}

		if err := fs.DB.Unscoped().Delete(&toBeDeletedFiles).Error; err != nil {
			return err
		}
	}

	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)
		for _, file := range toBeDeletedFiles {
			if len(toBeDeletedFiles) > 0 {
				obj := minio.ObjectInfo{
					Key: file.FileCode,
				}
				objectsCh <- obj
			}
		}
	}()

	for err := range bc.Client.RemoveObjects(bc.Context, bc.Bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return err.Err
		}
	}

	log.Printf("Deleting folder %s (%s)\n", folder.Name, folder.Code)

	deletedObjects.DeletedFolders = append(deletedObjects.DeletedFolders, folder.Code)

	if err := fs.DB.Unscoped().Delete(&folder).Error; err != nil {
		return err
	}

	return nil
}