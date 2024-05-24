package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

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

	path := form.Value["path"][0]
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid path",
		})
		return
	}

	// Remove trailing slash
	path = strings.TrimSuffix(path, "/")

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read file",
		})
		return
	}

	fileName := fmt.Sprintf("/%s/%s", path, file.Filename)

	// UPLOAD FILE RECORD TO RDBMS
	// Create new File record in rbdms
	newFile := models.File{
		UserID: userClaim.ID,
		FileName: file.Filename,
		FileSize: uint(file.Size),
		FileType: file.Header.Get("Content-Type"),
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

	_, err = minioClient.PutObject(ctx, user.MinioBucket,  fileName, uploadedFile, file.Size, minio.PutObjectOptions{ContentType: file.Header.Get("Content-Type")})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to upload file",
		})
		log.Println(err.Error())
		return
	}

	c.Status(http.StatusCreated)
}
