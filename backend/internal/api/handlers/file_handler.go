package handlers

import (
	"errors"
	"log"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type FileHandler struct {
	FileService *services.FileService
}

func NewFileHandler(fileService *services.FileService) *FileHandler {
	return &FileHandler{
		FileService: fileService,
	}
}

func (h *FileHandler) FileFavorites(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	files, err := h.FileService.ListFavoriteFiles(userClaim.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, files)
}

func (h *FileHandler) FileTrashCan(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	files, err := h.FileService.ListTrashCanFiles(userClaim.ID)
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, files)
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

func (fh *FileHandler) FilePatch(c *gin.Context) {
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	fileID := c.Param("fileID")

	validate := validator.New()

	var fileUpdateBody models.FilePatchBody

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

	intFileID, err := strconv.Atoi(fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid file ID",
		})
		return
	}

	file, err := fh.FileService.PatchFile(userClaim.ID, uint(intFileID), fileUpdateBody)
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

func (fh *FileHandler) FileDownload(c *gin.Context) {
	fileCode := c.Param("fileCode")
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)

	presignedURL, err := fh.FileService.GetPresignedURL(userClaim.ID, fileCode)
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

	proxy := httputil.NewSingleHostReverseProxy(presignedURL)
	proxy.Director = func(req *http.Request) {
		req.Host = presignedURL.Host
		req.URL.Scheme = presignedURL.Scheme
		req.URL.Host = presignedURL.Host
		req.URL.Path = presignedURL.Path
		req.URL.RawQuery = presignedURL.RawQuery
	}

	proxy.ServeHTTP(c.Writer, c.Request)
}

func (h *FileHandler) FileThumbnail(c *gin.Context) {
	fileCode := c.Param("fileCode")
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	isDeleted := c.DefaultQuery("deleted", "false") == "true"

	thumbnailService := services.NewThumbnailService(h.FileService.DB, h.FileService.BucketClient)

	thumbnail, err := thumbnailService.GetThumbnail(fileCode, userClaim.ID, isDeleted)
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
