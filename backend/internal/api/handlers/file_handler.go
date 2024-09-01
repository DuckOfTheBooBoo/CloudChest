package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type FileHandler struct {
	FileService *services.FileService
}

func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		FileService: fileService,
	}
}

func (h *FileHandler) FileList(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	isTrashCan := c.DefaultQuery("trashCan", "false") == "true"
	isFavorite := c.DefaultQuery("favorite", "false") == "true"

	files, err := h.FileService.ListFiles(userClaim.ID, isTrashCan, isFavorite)

	if err != nil {
		if errors.Is(err, &apperr.ServerError{}) {
			c.Status(http.StatusInternalServerError)
			return
		} else if errors.Is(err, &apperr.InvalidParamError{}) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Cannot fetch trash can and favorite at the same time.",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"files": files,
	})
}

func FileDelete(c *gin.Context) {
	ctx := context.Background()
	db := c.MustGet("db").(*gorm.DB)
	minioClient := c.MustGet("minio").(*minio.Client)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	fileID := c.Param("fileID")

	var user models.User
	if err := db.Where("id = ?", userClaim.ID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	isTrashDelete := c.DefaultQuery("trash", "true") == "true"
	isPruneAll := c.DefaultQuery("pruneAll", "false") == "true"

	if fileID == "" && !isTrashDelete {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File ID is required"})
		return
	}

	if isTrashDelete && isPruneAll {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot use trash can and prune all at the same time."})
		return
	}

	// PERMANENT DELETE
	if !isTrashDelete {
		var file models.File
		if err := db.Unscoped().Where("id = ? AND user_id = ?", fileID, userClaim.ID).Preload("Thumbnail").First(&file).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		if err := deleteFile(db, minioClient, ctx, file, user.MinioBucket); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
			log.Println("Error during file deletion: ", err.Error())
			return
		}
		c.Status(http.StatusOK)
		return
	}

	// SOFT DELETE
	var file models.File
	if err := db.Where("id = ? AND user_id = ?", fileID, userClaim.ID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	if err := db.Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
		log.Println("Error during soft deletion: ", err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func FileDeleteAll(c *gin.Context) {
	ctx := context.Background()
	db := c.MustGet("db").(*gorm.DB)
	minioClient := c.MustGet("minio").(*minio.Client)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	var user models.User
	if err := db.Where("id = ?", userClaim.ID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if err := deleteAllFilesForUser(db, minioClient, ctx, userClaim.ID, user.MinioBucket); err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println("Error during pruning all files: ", err.Error())
		return
	}

	c.Status(http.StatusOK)
}

func deleteAllFilesForUser(db *gorm.DB, minioClient *minio.Client, ctx context.Context, userID uint, bucket string) error {
	var deletedFiles []models.File
	if err := db.Unscoped().Preload("Thumbnail").Where("user_id = ? AND deleted_at IS NOT NULL", userID).Find(&deletedFiles).Error; err != nil {
		return err
	}

	for _, file := range deletedFiles {
		if err := deleteFile(db, minioClient, ctx, file, bucket); err != nil {
			return err
		}
	}
	return nil
}

func deleteFile(db *gorm.DB, minioClient *minio.Client, ctx context.Context, file models.File, bucket string) error {
	// Delete Thumbnail
	if file.Thumbnail != nil {
		if err := minioClient.RemoveObject(ctx, bucket, file.Thumbnail.FilePath, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("failed to delete thumbnail from MinIO: %w", err)
		}
		if err := db.Unscoped().Delete(&file.Thumbnail).Error; err != nil {
			return fmt.Errorf("failed to delete thumbnail from database: %w", err)
		}
	}

	// Delete HLS segments and master playlist
	if file.IsPreviewable && strings.HasPrefix(file.FileType, "video/") {
		if err := deleteHLSFiles(minioClient, ctx, bucket, file.FileCode); err != nil {
			return fmt.Errorf("failed to delete HLS files: %w", err)
		}
	}

	// DELETE FROM MINIO
	if err := minioClient.RemoveObject(ctx, bucket, file.FileCode, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	// DELETE FROM DB
	if err := db.Unscoped().Delete(&file).Error; err != nil {
		return fmt.Errorf("failed to delete file from database: %w", err)
	}

	return nil
}

func deleteHLSFiles(minioClient *minio.Client, ctx context.Context, bucket, fileCode string) error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range minioClient.ListObjects(ctx, bucket, minio.ListObjectsOptions{
			Prefix:    "hls/" + fileCode,
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Println("Error listing HLS files: ", object.Err)
				continue
			}
			if strings.HasSuffix(object.Key, ".m3u8") || strings.HasSuffix(object.Key, ".ts") {
				objectsCh <- object
			}
		}
	}()

	for rErr := range minioClient.RemoveObjects(ctx, bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if rErr.Err != nil {
			return fmt.Errorf("error detected during HLS file deletion: %w", rErr.Err)
		}
	}
	return nil
}

func FileUpdate(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	fileID := c.Param("fileID")

	var user models.User
	err := db.Where("id = ?", userClaim.ID).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	var fileUpdateBody struct {
		FileName   string `validate:"required" json:"file_name"`
		IsFavorite bool   `validate:"boolean" json:"is_favorite"`
		Restore    bool   `validate:"boolean" json:"is_restore"`
	}

	validate := validator.New()

	if err := c.BindJSON(&fileUpdateBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No request body (JSON) included.",
		})
		return
	}

	if err := validate.Struct(fileUpdateBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// {
	// 	"FileName": "error.log",
	// 	"IsFavorite": false,
	// 	"Restore": false
	// }

	// Find file
	var file models.File
	if !fileUpdateBody.Restore {
		if err := db.Where("id = ? AND user_id = ?", fileID, user.ID).First(&file).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			log.Println(err.Error())
			return
		}
	} else {
		if err := db.Unscoped().Where("id = ? AND user_id = ?", fileID, user.ID).First(&file).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			log.Println(err.Error())
			return
		}
	}

	if fileUpdateBody.FileName != file.FileName {
		file.FileName = fileUpdateBody.FileName
	}

	file.IsFavorite = fileUpdateBody.IsFavorite

	if fileUpdateBody.Restore {
		if err := db.Unscoped().Model(&file).Update("deleted_at", nil).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to restore file",
			})
			log.Println(err.Error())
			return
		}
	}

	if err := db.Save(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update file",
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, file)
}

func FilePatch(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	fileID := c.Param("fileID")

	validate := validator.New()

	var fileUpdateBody struct {
		FileName         string `json:"file_name"`
		FolderCode string `validate:"ascii" json:"folder_code"`
		IsFavorite       bool   `validate:"boolean" json:"is_favorite"`
		Restore          bool   `validate:"boolean" json:"is_restore"`
	}

	if err := c.BindJSON(&fileUpdateBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No request body (JSON) included.",
		})
		return
	}
	
	if err := validate.Struct(fileUpdateBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		log.Println(fileUpdateBody)
		return
	}

	// Find file
	var file models.File
	if !fileUpdateBody.Restore {
		if err := db.Where("id = ? AND user_id = ?", fileID, userClaim.ID).First(&file).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			log.Println(err.Error())
			return
		}
	} else {
		if err := db.Unscoped().Where("id = ? AND user_id = ?", fileID, userClaim.ID).First(&file).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			log.Println(err.Error())
			return
		}
	}

	if fileUpdateBody.FileName != "" {
		file.FileName = fileUpdateBody.FileName
	}

	if fileUpdateBody.FolderCode != "" {
		var parentFolder models.Folder
		err := db.Where("user_id = ? AND code = ?", userClaim.ID, fileUpdateBody.FolderCode).First(&parentFolder).Error;
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "Parent folder not found",
				})
				log.Println(err.Error())
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Printf("Failed to find parent folder: %v", err)
			return
		} 

		file.FolderID = parentFolder.ID
	}

	if file.IsFavorite != fileUpdateBody.IsFavorite {
		file.IsFavorite = fileUpdateBody.IsFavorite
	}

	if fileUpdateBody.Restore {
		if err := db.Unscoped().Model(&file).Update("deleted_at", nil).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to restore file",
			})
			log.Println(err.Error())
			return
		}
	}

	if err := db.Save(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update file",
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, file)
}

func FileDownload(c *gin.Context) {
	ctx := context.Background()
	minioClient := c.MustGet("minio").(*minio.Client)
	fileID := c.Param("fileID")
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	// Get user bucket name
	var user models.User
	if err := db.Where("id = ?", userClaim.ID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	// Get file
	var file models.File
	if err := db.Where("id = ? AND user_id = ?", fileID, user.ID).First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\""+file.FileName+"\"")

	charsetParam := ""
	if strings.HasPrefix(file.FileType, "text/") {
		charsetParam = "; charset=utf-8"
	}

	reqParams.Set("response-content-type", file.FileType+charsetParam)

	presignedURL, err := minioClient.PresignedGetObject(ctx, user.MinioBucket, file.FileCode, time.Second*24*60*60, reqParams)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, presignedURL)
}

func FileThumbnail(c *gin.Context) {
	ctx := context.Background()
	minioClient := c.MustGet("minio").(*minio.Client)
	fileID := c.Param("fileID")
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	// Get user bucket name
	var user models.User
	if err := db.Where("id = ?", userClaim.ID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	// Get file
	var file models.File
	if err := db.Where("id = ? AND user_id = ?", fileID, user.ID).Preload("Thumbnail").First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	if file.Thumbnail == nil {
		if strings.HasPrefix(file.FileType, "image/") {
			c.JSON(http.StatusAccepted, gin.H{
				"message": "File's thumbnail is being processed",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "File is not an image or a video",
			})
		}
		return
	}

	thumbnail, err := minioClient.GetObject(ctx, user.MinioBucket, file.Thumbnail.FilePath, minio.GetObjectOptions{})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	defer thumbnail.Close()

	// Read thumbnail's info
	info, err := thumbnail.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.DataFromReader(http.StatusOK, info.Size, info.ContentType, thumbnail, nil)
}
