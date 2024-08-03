package controllers

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	// "strings"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/gofrs/uuid/v5"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func FileList(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.DefaultQuery("code", "/")
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

	var files []models.File
	if err := db.Where("user_id = ? AND code = ?", user.ID, folderCode).Find(&files).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	var parentFolder models.Folder
	if err := db.Where("user_id = ? AND code = ?", user.ID, folderCode).First(&parentFolder).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	var subFolders []models.Folder
	// Generate parent and child folders
	if err := db.Where("user_id = ? AND parent_id = ?", user.ID, parentFolder.ID).Find(&subFolders).Error; err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files":   files,
		"folders": subFolders,
	})
}

func FileUpload(c *gin.Context) {
	ctx := context.Background()
	minioClient := c.MustGet("minio").(*minio.Client)
	folderCode := c.DefaultQuery("code", "/")
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
	if folderCode == "/" {
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
			UserID:      userClaim.ID,
			FolderID: parentFolder.ID,
			FileName:    file.Filename,
			FileCode:    fileCode.String(),
			FileSize:    uint(file.Size),
			FileType:    file.Header.Get("Content-Type"),
			IsFavorite:  false,
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

		fileExt := utils.GetFileExtension(file.Filename)
		filePath := "/" + fileCode.String() + "." + fileExt

		_, err = minioClient.PutObject(ctx, user.MinioBucket, filePath, uploadedFile, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})

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
		fileCode, err := uuid.NewV4()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		newFile := models.File{
			UserID:      userClaim.ID,
			FolderID: parentFolder.ID,
			FileName:    file.Filename,
			FileCode:    fileCode.String(),
			FileSize:    uint(file.Size),
			FileType:    file.Header.Get("Content-Type"),
			IsFavorite:  false,
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

		fileExt := utils.GetFileExtension(file.Filename)
		filePath := "/" + fileCode.String() + "." + fileExt

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

		fileExt := utils.GetFileExtension(file.FileName)

		// DELETE FROM MINIO
		err = minioClient.RemoveObject(ctx, user.MinioBucket, "/" + file.FileCode + "." + fileExt, minio.RemoveObjectOptions{})

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
			"error": "Failed to create new folder",
		})
		log.Println(err.Error())
		return
	}

	// Split path into parts
	pathParts := strings.Split(path, "/")
	// Get the latest part
	// latestPart := pathParts[len(pathParts)-1]
	// Remove the latest part from the path
	pathParts = pathParts[:len(pathParts)-1]
	parentPath := strings.Join(pathParts[:], "/") 

	if len(parentPath) == 0 {
		parentPath = "/"
	}

	// if err := db.Create(&newFolderChildRecord).Error; err != nil {
	// 	c.Status(http.StatusInternalServerError)
	// 	log.Println(err.Error())
	// 	return
	// }

	c.JSON(http.StatusCreated, gin.H{
		"path": path,
	})
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
		FileName   string `validate:"required"`
		IsFavorite bool   `validate:"boolean"`
		Restore bool `validate:"boolean"`
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
		c.Status(http.StatusMethodNotAllowed)
		return
	}

	file.FileName = fileUpdateBody.FileName
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

	c.JSON(http.StatusOK, gin.H{
		"file": file,
	})
}
