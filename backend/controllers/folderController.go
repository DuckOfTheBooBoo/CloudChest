package controllers

import (
	"errors"
	"log"
	"net/http"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

func FolderList(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")

	var parentFolder models.Folder
	if folderCode == "" {
		if err := db.Where("user_id = ? AND (code IS NULL OR code = '')", userClaim.ID).Find(&parentFolder).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		var subFolders []models.Folder
		if err := db.Where("user_id = ? AND parent_id = ?", userClaim.ID, parentFolder.ID).Find(&subFolders).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.Status(http.StatusNotFound)
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		c.JSON(http.StatusOK, subFolders)
		return
	}

	if err := db.Where("user_id = ? AND code = ?", userClaim.ID, folderCode).Preload("ChildFolders").Find(&parentFolder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.Status(http.StatusNotFound)
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, parentFolder.ChildFolders)
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
	if parentFolderCode != "" {
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
		UserID: userClaim.ID,
		ParentID: &parentFolder.ID,
		Name:   folderBody.FolderName,
		Code: newFolderCode,
	}

	if err := db.Create(&newFolder).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create folder.",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"folder": newFolder,
	})
}