package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func FileList(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
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
			"error": "Cannot fetch trash can and favorite at the same time.",
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

	// Trash can
	var trashedFiles []models.File
	if err := db.Unscoped().Where("user_id = ? AND deleted_at IS NOT NULL", user.ID).Find(&trashedFiles).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": trashedFiles,
	})
}

func FileDelete(c *gin.Context) {
	ctx := context.Background()
	db := c.MustGet("db").(*gorm.DB)
	minioClient := c.MustGet("minio").(*minio.Client)
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

	var file models.File

	// Check if user wants to trash or permanently delete
	isTrashDelete := c.DefaultQuery("trash", "true") == "true"
	// PERMANENT DELETE
	if !isTrashDelete {
		if err = db.Unscoped().Where("id = ? AND user_id = ?", fileID, userClaim.ID).Preload("Thumbnail").First(&file).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}

		if file.Thumbnail != nil {
			// DELETE FROM MINIO
			if err := minioClient.RemoveObject(ctx, user.MinioBucket, file.Thumbnail.FilePath, minio.RemoveObjectOptions{}); err != nil {
				c.Status(http.StatusInternalServerError)
				log.Println("Failed to delete thumbnail from MinIO: ", err.Error())
				return
			}

			if err := db.Unscoped().Model(&file).Association("Thumbnail").Unscoped().Clear(); err != nil {
				c.Status(http.StatusInternalServerError)
				log.Println("Failed to delete thumbnail from database: ", err.Error())
				return
			}
		}

		// DELETE FROM MINIO
		if err := minioClient.RemoveObject(ctx, user.MinioBucket, "/"+file.FileCode, minio.RemoveObjectOptions{}); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete file",
			})
			log.Println(err.Error())
			return
		}

		err = db.Unscoped().Delete(&file).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete file",
			})
			log.Println(err.Error())
			return
		}

		c.Status(http.StatusOK)
		return
	}

	if err := db.Where("id = ? AND user_id = ?", fileID, userClaim.ID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	// Soft delete the file
	if err := db.Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file",
		})
		log.Println(err.Error())
		return
	}

	c.Status(http.StatusOK)
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