package handlers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"
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

func (h *FileHandler) FileDelete(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	fileID := c.Param("fileID")
	isTrashDelete := c.DefaultQuery("trash", "true") == "true"

	intFileID, err := strconv.Atoi(fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	// PERMANENT DELETE
	if !isTrashDelete {
		if err := h.FileService.DeleteFilePermanent(userClaim.ID, uint(intFileID)); err != nil {
			if errors.Is(err, &apperr.NotFoundError{}) {
				c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
				return
			}

			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
		return
	}

	// SOFT DELETE
	if err := h.FileService.DeleteFileTemp(userClaim.ID, uint(intFileID)); err != nil {
		if errors.Is(err, &apperr.NotFoundError{}) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		c.Status(http.StatusInternalServerError)
		return
	}
	c.Status(http.StatusOK)
}

func (h *FileHandler) FileDeleteAll(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	if err := h.FileService.EmptyTrashCan(userClaim.ID); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (fh *FileHandler) FileUpdate(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	fileID := c.Param("fileID")

	var fileUpdateBody models.FileUpdateBody
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

	intFileID, err := strconv.Atoi(fileID)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file ID",
		})
		return
	}

	file, err := fh.FileService.UpdateFile(userClaim.ID, uint(intFileID), fileUpdateBody)
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

func (h *FileHandler) FileThumbnail(c *gin.Context) {
	fileCode := c.Param("fileCode")
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	thumbnailService := services.NewThumbnailService(h.FileService.DB, h.FileService.BucketClient)

	thumbnail, err := thumbnailService.GetThumbnail(fileCode, userClaim.ID)
	if err != nil {
		if errors.Is(err, &apperr.InvalidParamError{}) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		} else if errors.Is(err, &apperr.NotFoundError{}) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		} else if errors.Is(err, &apperr.ResourceNotReadyError{}) {
			c.JSON(http.StatusAccepted, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	if thumbnail != nil {
		defer thumbnail.Close()
	}

	// Read thumbnail's info
	info, err := thumbnail.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.DataFromReader(http.StatusOK, info.Size, info.ContentType, thumbnail, nil)
}
