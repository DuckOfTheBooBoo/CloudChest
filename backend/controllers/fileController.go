package controllers

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"path/filepath"
	// "strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func FileList(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	path := c.Query("path")

	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No path provided",
		})
		return
	}

	var user models.User
	err := db.First(&user, "id = ?", userClaim.ID).Error

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	var files []models.File
	if err := db.Where("user_id = ? AND dir_path = ?", user.ID, path).Find(&files).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	var folderChildren []models.FolderChild
	// Generate parent and child folders
	if err := db.Where("user_id = ? AND parent = ?", user.ID, path).Find(&folderChildren).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	var folders []models.Folder
	for _, folderChild := range folderChildren {
		if folderChild.Child != "" {
			folder := models.Folder{
				DirName:   folderChild.Child,
				CreatedAt: folderChild.CreatedAt,
				UpdatedAt: folderChild.UpdatedAt,
			}

			folders = append(folders, folder)
		}
	}

	// Filter slice to match path provided by request body
	files = utils.FilterSlice(files, func(file models.File) bool {
		return filepath.Dir(file.StoragePath) == path
	})

	c.JSON(http.StatusOK, gin.H{
		"files":   files,
		"folders": folders,
	})
}

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

		if filepath.Dir(fileName) != "" || filepath.Dir(fileName) != "/" {
			// Create a parent dir -> child dir mapping
			log.Println("CREATING PARENT CHILD MAP")
			parentChildDir := utils.GenerateParentChildDir(userClaim.ID, filepath.Dir(fileName))

			for _, parentChild := range parentChildDir {
				db.Create(&parentChild)
			}
		}

		// UPLOAD FILE RECORD TO RDBMS
		// Create new File record in rbdms
		newFile := models.File{
			UserID:      userClaim.ID,
			FileName:    file.Filename,
			FileSize:    uint(file.Size),
			FileType:    file.Header.Get("Content-Type"),
			DirPath:     filepath.Dir(fileName),
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
	baseFilePath := filepath.Dir(newFiles[0].StoragePath)

	for _, file := range files {
		fileName := filepath.Join(path, file.Filename)
		newFile := models.File{
			UserID:      userClaim.ID,
			FileName:    file.Filename,
			FileSize:    uint(file.Size),
			FileType:    file.Header.Get("Content-Type"),
			DirPath:     filepath.Dir(fileName),
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

	if baseFilePath != "" || baseFilePath != "/" {
		// Create a parent dir -> child dir mapping
		log.Println("CREATING PARENT CHILD MAP")
		parentChildDir := utils.GenerateParentChildDir(userClaim.ID, baseFilePath)

		for _, parentChild := range parentChildDir {
			db.Create(&parentChild)
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

func FileNewPath(c *gin.Context) {
	ctx := context.Background()
	db := c.MustGet("db").(*gorm.DB)
	minioClient := c.MustGet("minio").(*minio.Client)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	path := c.DefaultQuery("path", "/")

	var user models.User
	err := db.Where("id = ?", userClaim.ID).First(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// To create a new path, we could make empty file and upload it to the desired path, then delete it
	emptyFile := []byte{}
	emptyFileName := ".newPath"

	reader := bytes.NewReader(emptyFile)

	if _, err := io.Copy(io.Discard, reader); err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	// file path
	emptyFilePath := filepath.Join(path, emptyFileName)

	// Upload empty file
	_, err = minioClient.PutObject(ctx, user.MinioBucket, emptyFilePath, reader, 0, minio.PutObjectOptions{ContentType: "text/plain"})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload file",
		})
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"path": path,
	})
}
