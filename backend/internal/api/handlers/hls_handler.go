package handlers

import (
	"net/http"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/gin-gonic/gin"
)

type HLSHandler struct {
	HLSService *services.HLSService
}

func NewHLSHandler(hlsService *services.HLSService) *HLSHandler {
	return &HLSHandler{
		HLSService: hlsService,
	}
}

func (h *HLSHandler) ServeMasterPlaylist(c *gin.Context) {
	fileCode := c.Param("fileCode")

	masterPlaylist, size, err := h.HLSService.GetMasterPlaylist(fileCode)
	if err != nil {
		switch err.(type) {
			case *apperr.NotFoundError:
				c.Status(http.StatusNotFound)
				return
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	c.DataFromReader(http.StatusOK, *size, "application/vnd.apple.mpegurl", masterPlaylist, nil)
}

func (h *HLSHandler) ServeSegment(c *gin.Context) {
	fileCode := c.Param("fileCode")
	segmentNum := c.Param("segmentNumber")

	segment, size, err := h.HLSService.GetSegment(fileCode, segmentNum)
	if err != nil {
		switch err.(type) {
			case *apperr.NotFoundError:
				c.Status(http.StatusNotFound)
				return
		}

		c.Status(http.StatusInternalServerError)
		return
	}

	c.DataFromReader(http.StatusOK, *size, "video/MP2T", segment, nil)
}