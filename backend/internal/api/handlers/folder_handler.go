package handlers

import (
	"errors"
	"log"
	"net/http"

	// "sort"
	"strings"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)


type FolderHandler struct {
	FolderService *services.FolderService
}

func NewFolderHandler(fs *services.FolderService) *FolderHandler {
	return &FolderHandler{
		FolderService: fs,
	}
}

func (fh *FolderHandler) FolderList(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")
	
	var folderResp *models.FolderResponse
	var err error

	// Dirty workaround but meh
	if folderCode == "favorite" {
		folderResp, err = fh.FolderService.ListFavoriteFolders(userClaim.ID) 
	} else if folderCode == "trashcan" {
		folderResp, err = fh.FolderService.ListTrashFolders(userClaim.ID)
	} else {
		folderResp, err = fh.FolderService.ListFolders(userClaim.ID, folderCode)
	}

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

func (fh *FolderHandler) FolderCreate(c *gin.Context) {
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

	newFolder, err := fh.FolderService.CreateFolder(folderBody.FolderName, parentFolderCode, userClaim.ID)
	if err != nil {
		if errors.Is(err, &apperr.NotFoundError{}) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"folder": newFolder,
	})
}

func (fh *FolderHandler) FolderContentsCreate(c *gin.Context) {
	folderCode := c.Param("code")
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read file",
		})
		return
	}

	newFile, fileBytes, err := fh.FolderService.UploadFile(userClaim.ID, folderCode, file)
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

	c.JSON(http.StatusCreated, newFile)

	if strings.HasPrefix(newFile.FileType, "image/") || strings.HasPrefix(newFile.FileType, "video/") {
		fh.FolderService.PostUploadProcess(newFile, fileBytes)
	}
}

func (fh *FolderHandler) FolderContents(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	folderCode := c.Param("code")

	files, err := fh.FolderService.FetchFolderFiles(userClaim.ID, folderCode)
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

	folder, err := fh.FolderService.PatchFolder(userClaim.ID, folderCode, folderUpdateBody)
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

func (fh *FolderHandler) FolderDelete(c *gin.Context) {
	folderCode := c.Param("code")
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	trash := c.DefaultQuery("trash", "true") == "true"

	if trash {
		if err := fh.FolderService.DeleteFolderTemp(folderCode, userClaim.ID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			}

			c.Status(http.StatusInternalServerError)
			log.Println(err.Error())
			return
		}

		c.Status(http.StatusOK)
		return
	}

	deletedObjects, err := fh.FolderService.DeleteFolderPermanent(folderCode, userClaim.ID);
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, deletedObjects)
}
