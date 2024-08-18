package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/jobs"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

const (
	MAX_PREVIEWABLE_VIDEO_SIZE = 150 * 1000 * 1000
)

func FolderList(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")

	var parentFolder models.Folder
	if folderCode == "root" {
		if err := db.Where("user_id = ? AND (code IS NULL OR code = '')", userClaim.ID).Preload("ChildFolders").Find(&parentFolder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	} else {
		if err := db.Where("user_id = ? AND code = ?", userClaim.ID, folderCode).Preload("ChildFolders").Find(&parentFolder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}
	}
	// Generate hierarchy
	var currentParent *models.Folder = &parentFolder
	var hierarchies []models.FolderHierarchy
	for currentParent.ParentID != nil {
		var parent models.Folder
		if err := db.Where("id = ?", *currentParent.ParentID).Find(&parent).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
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

	c.JSON(http.StatusOK, gin.H{
		"folders":     parentFolder.ChildFolders,
		"hierarchies": hierarchies,
	})
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
						case <- ctx.Done():
							log.Println("Thumbnail generation cancelled due to error in writing temp file.")
							return
						
						default:
							filePath := <- tempThumbFilePathChan
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
						case <- ctx.Done():
							log.Println("HLS process cancelled due to error in writing temp file.")
							return
						
						default:
							filePath := <- tempHLSbFilePathChan
							jobs.ProcessHLS(filePath, ctx, minioClient, db, newFile, user.MinioBucket)
						}
					}(jobCtx)
				}

				// Remove temp file
				wg.Wait()
				filePath := <- tempThumbFilePathChan
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

func FolderContents(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")
	isTrashCan := c.DefaultQuery("trashCan", "false") == "true"
	isFavorite := c.DefaultQuery("favorite", "false") == "true"

	var user models.User
	err := db.First(&user, "id = ?", userClaim.ID).Error

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	if isTrashCan && isFavorite {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot use trash can and favorite at the same time.",
		})
		return
	}

	if isFavorite {
		var favoriteFiles []models.File
		if err := db.Where("user_id = ? AND is_favorite = ?", user.ID, true).Find(&favoriteFiles).Error; err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"files": favoriteFiles,
		})
		return
	}

	if isTrashCan {
		var trashedFiles []models.File
		if err := db.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", user.ID).Find(&trashedFiles).Error; err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"files": trashedFiles,
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

	var files []models.File
	if err := db.Where("user_id = ? AND folder_id = ?", user.ID, parentFolder.ID).Find(&files).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, files)
}
