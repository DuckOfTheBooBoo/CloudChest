package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	// "sort"
	"strings"
	"sync"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/jobs"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MAX_PREVIEWABLE_VIDEO_SIZE = 150 * 1000 * 1000
)

type FolderHandler struct {
	folderService *services.FolderService
}

func NewFolderHandler(fs *services.FolderService) *FolderHandler {
	return &FolderHandler{
		folderService: fs,
	}
}

func (fh *FolderHandler) FolderList(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")



	folderResp, err := fh.folderService.ListFolders(userClaim.ID, folderCode)
	if err != nil {
		switch e := err.(type) {
			case *apperr.NotFoundError:
				c.JSON(http.StatusNotFound, gin.H{
					"error": e.Error(),
				})
				return
			case *apperr.ServerError:
				c.Status(http.StatusInternalServerError)
				return
		}
	}

	c.JSON(http.StatusOK, folderResp)
}

func FolderCreate(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	parentFolderCode := c.Param("code")
	validate := validator.New()

	var folderBody struct {
		FolderName string `json:"folder_name" validate:"required,ascii"`
	}

	if err := c.BindJSON(&folderBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No request body (JSON) included.",
		})
		return
	}

	if err := validate.Struct(folderBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	newFolderCode, err := gonanoid.New()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate folder code.",
		})
		return
	}

	// Fetch parent folder
	var parentFolder models.Folder
	// Query by parent folder code
	if parentFolderCode != "root" {
		if err := db.Where("user_id = ? AND code = ?", userClaim.ID, parentFolderCode).First(&parentFolder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	} else {
		// Query by user parent folder
		if err := db.Where("user_id = ? AND (code IS NULL OR code = '')", userClaim.ID).First(&parentFolder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	}
	newFolder := models.Folder{
		UserID:   userClaim.ID,
		ParentID: &parentFolder.ID,
		Name:     folderBody.FolderName,
		Code:     newFolderCode,
	}

	if err := db.Create(&newFolder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create folder.",
		})
		return
	}

	parentFolder.HasChild = true

	if err := db.Save(&parentFolder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update parent folder.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"folder": newFolder,
	})
}

func FolderContentsCreate(c *gin.Context) {
	ctx := context.Background()
	minioClient := c.MustGet("minio").(*minio.Client)
	folderCode := c.Param("code")
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	// Check if upload is uploading multiple files
	isMultipleUploads := c.DefaultQuery("multiple", "false") == "true"

	// Read user from database
	var user models.User
	err := db.First(&user, "id = ?", userClaim.ID).Error

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	form, err := c.MultipartForm()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var parentFolder models.Folder
	if folderCode == "root" {
		if err := db.Where("user_id = ? AND (code IS NULL OR code = '')", user.ID).First(&parentFolder).Error; err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	} else {
		if err := db.Where("user_id = ? AND code = ?", user.ID, folderCode).First(&parentFolder).Error; err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	}

	if !isMultipleUploads {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read file",
			})
			return
		}

		fileCode, err := uuid.NewV4()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		// UPLOAD FILE RECORD TO RDBMS
		// Create new File record in rbdms
		newFile := models.File{
			UserID:     userClaim.ID,
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
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to open file",
			})
			log.Println(err.Error())
			return
		}

		// Stuf happens, uploadedFile have a body length of 0? It went well until I pass it into minioClient for uploading to the bucket. The current workaround is to convert it into []byte regardless its mime type
		// Prepare []byte of the uploded file in case its a media (image or video) file. Else let it collected by garbage collector
		var uploadedFileBytes []byte
		uploadedFileBytes, err = io.ReadAll(uploadedFile)
		if err != nil {
			log.Printf("Error while reading uploaded file: %v", err)
			return
		}
		uploadedFile.Close()

		filePath := "/" + fileCode.String()

		// Use transaction, PutObject to minio could lead to an error. If it does, we can't let any changes happen in the database
		err = db.Transaction(func(tx *gorm.DB) error {
			// Upload the file to minio first
			_, err = minioClient.PutObject(ctx, user.MinioBucket, filePath, bytes.NewReader(uploadedFileBytes), file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})
			if err != nil {
				return fmt.Errorf("error while uploading file to MinIO: %v", err)
			}

			if err := db.Create(&newFile).Error; err != nil {
				return fmt.Errorf("error while creating file in database: %v", err)
			}

			return nil
		})

		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create file",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upload file",
			})
			log.Println(err.Error())
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"file": newFile,
		})

		if strings.HasPrefix(newFile.FileType, "image/") || strings.HasPrefix(newFile.FileType, "video/") {
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

					result, err := jobs.WriteTempFile(newFile, uploadedFileBytes)
					if err != nil {
						cancel()
						log.Println(err)
						return
					}
					tempThumbFilePathChan <- result
					tempHLSbFilePathChan <- result
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
						jobs.GenerateThumbnail(ctx, filePath, minioClient, db, newFile, user.MinioBucket)
					}

				}(jobCtx)

				// Process HLS file (video only)
				if strings.HasPrefix(newFile.FileType, "video/") && newFile.FileSize <= MAX_PREVIEWABLE_VIDEO_SIZE {
					log.Printf("Processing %s for HLS", newFile.FileName)
					wg.Add(1)
					go func(ctx context.Context) {
						defer wg.Done()
						select {
						case <-ctx.Done():
							log.Println("HLS process cancelled due to error in writing temp file.")
							return

						default:
							filePath := <-tempHLSbFilePathChan
							jobs.ProcessHLS(filePath, ctx, minioClient, db, newFile, user.MinioBucket)
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

		return
	}

	files := form.File["files"]
	var newFiles []models.File

	for _, file := range files {
		fileCode, err := uuid.NewV4()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		newFile := models.File{
			UserID:     userClaim.ID,
			FolderID:   parentFolder.ID,
			FileName:   file.Filename,
			FileCode:   fileCode.String(),
			FileSize:   uint(file.Size),
			FileType:   file.Header.Get("Content-Type"),
			IsFavorite: false,
		}

		newFiles = append(newFiles, newFile)

		// UPLOAD FILE TO MINIO
		// Read the file
		uploadedFile, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to open file " + file.Filename,
			})
			log.Println(err.Error())
			return
		}
		defer uploadedFile.Close()

		filePath := "/" + fileCode.String()

		_, err = minioClient.PutObject(ctx, user.MinioBucket, filePath, uploadedFile, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upload file",
			})
			log.Println(err.Error())
			return
		}
	}

	// Upload newFiles to rdbms
	if err := db.Create(&newFiles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create files",
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"files": newFiles,
	})
}

func (fh *FolderHandler) FolderContents(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")

	files, err := fh.folderService.FetchFolderFiles(userClaim.ID, folderCode)
	if err != nil {
		switch err := err.(type) {
			case *apperr.NotFoundError:
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			case *apperr.ServerError:
				c.Status(http.StatusInternalServerError)
				return
		}
	}

	c.JSON(http.StatusOK, files)
}

func (fh *FolderHandler) FolderPatch(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")

	validate := validator.New()

	var folderUpdateBody models.FolderUpdateBody

	if err := c.BindJSON(&folderUpdateBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No request body (JSON) included.",
		})
		return
	}

	if err := validate.Struct(folderUpdateBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		log.Println(folderUpdateBody)
		return
	}

	folder, err := fh.folderService.PatchFolder(userClaim.ID, folderCode, folderUpdateBody)
	if err != nil {
		switch err := err.(type) {
			case *apperr.NotFoundError:
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			case *apperr.ServerError:
				c.Status(http.StatusInternalServerError)
				return
		}
	}

	c.JSON(http.StatusOK, folder)
}

func loadFolders(db *gorm.DB, folder *models.Folder) error {
	// Preload immediate child folders
	if err := db.Preload(clause.Associations).Find(&folder).Error; err != nil {
		return err
	}

	// Recursively load children of each child folder
	for i := range folder.ChildFolders {
		if err := loadFolders(db, folder.ChildFolders[i]); err != nil {
			return err
		}
	}

	return nil
}

func processFolder(ctx *context.Context, db *gorm.DB, minioClient *minio.Client, minioBucket string, folder *models.Folder) error {
	for _, child := range(folder.ChildFolders) {
		if err := processFolder(ctx, db, minioClient, minioBucket, child); err != nil {
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
		if err := db.Unscoped().Delete(&filesThumbnail).Error; err != nil {
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
	
		for err := range minioClient.RemoveObjects(*ctx, minioBucket, thumbObjCh, minio.RemoveObjectsOptions{}) {
			if err.Err != nil {
				return err.Err
			}
		}
	}

	// Delete files from DB
	if err := db.Unscoped().Delete(&toBeDeletedFiles).Error; err != nil {
		return err
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

	for err := range minioClient.RemoveObjects(*ctx, minioBucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return err.Err
		}
	}

	log.Printf("Deleting folder %s (%s)\n", folder.Name, folder.Code)
	if err := db.Unscoped().Delete(&folder).Error; err != nil {
		return err
	}

	return nil
}

func FolderDelete(c *gin.Context) {
	ctx := context.Background()
	minioClient := c.MustGet("minio").(*minio.Client)
	folderCode := c.Param("code")
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	trash := c.DefaultQuery("trash", "true") == "true"

	var user models.User
	if err := db.Where("id = ?", userClaim.ID).Find(&user).Error; err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// Fetch folder
	var targetFolder models.Folder
	query := db.Where("code = ? AND user_id = ?", folderCode, user.ID)

	if !trash {
		query = query.Unscoped()
	}

	if err := query.Find(&targetFolder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Folder not found",
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	
	if trash {
		if err := db.Delete(&targetFolder).Error; err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err)
			return
		}

		c.Status(http.StatusOK)
		return
	}

	if err := loadFolders(db, &targetFolder); err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	if err := processFolder(&ctx, db, minioClient, user.MinioBucket, &targetFolder); err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	c.Status(http.StatusOK)
}
