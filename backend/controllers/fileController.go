package controllers

import (
	"context"
	"log"
	"net/http"
	"path/filepath"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func FileUpload(c *gin.Context) {
	ctx := context.Background()
	minioClient := c.MustGet("minio").(*minio.Client)
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

	pathArray := form.Value["path"]
	if len(pathArray) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid path",
		})
		return
	}

	path := pathArray[0]

	if !isMultipleUploads {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Failed to read file",
			})
			return
		}

		fileName := filepath.Join(path, file.Filename)

		// UPLOAD FILE RECORD TO RDBMS
		// Create new File record in rbdms
		newFile := models.File{
			UserID:      userClaim.ID,
			FileName:    file.Filename,
			FileSize:    uint(file.Size),
			FileType:    file.Header.Get("Content-Type"),
			StoragePath: fileName,
		}

		err = db.Create(&newFile).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to create file",
			})
			log.Println(err.Error())
			return
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
		defer uploadedFile.Close()

		_, err = minioClient.PutObject(ctx, user.MinioBucket, fileName, uploadedFile, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})

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
		return
	}

	files := form.File["files"]
	var newFiles []models.File
	for _, file := range files {
		fileName := filepath.Join(path, file.Filename)
		newFile := models.File{
			UserID:      userClaim.ID,
			FileName:    file.Filename,
			FileSize:    uint(file.Size),
			FileType:    file.Header.Get("Content-Type"),
			StoragePath: fileName,
		}
		newFiles = append(newFiles, newFile)

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
		defer uploadedFile.Close()

		_, err = minioClient.PutObject(ctx, user.MinioBucket, fileName, uploadedFile, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to upload file",
			})
			log.Println(err.Error())
			return
		}
	}

	// Upload newFiles to rdbms
	err = db.Create(&newFiles).Error
	if err != nil {
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
	if !isTrashDelete {
		err = db.Unscoped().Where("id = ? AND user_id = ?", fileID, userClaim.ID).First(&file).Error

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File not found",
			})
			return
		}

		// DELETE FROM MINIO
		err = minioClient.RemoveObject(ctx, user.MinioBucket, file.StoragePath, minio.RemoveObjectOptions{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete file",
			})
			log.Println(err.Error())
			return
		}

		err = db.Delete(&file).Error
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

	err = db.Where("id = ? AND user_id = ?", fileID, userClaim.ID).First(&file).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File not found",
		})
		return
	}

	err = db.Delete(&file).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete file",
		})
		log.Println(err.Error())
		return
	}

	c.Status(http.StatusOK)
}
