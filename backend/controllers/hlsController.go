package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func ServeMasterPlaylist(c *gin.Context) {
	ctx := context.Background()
	fileCode := c.Param("fileCode")
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	db := c.MustGet("db").(*gorm.DB)
	minioClient := c.MustGet("minio").(*minio.Client)

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

	masterPlaylistPath := fmt.Sprintf("/hls/%s/%s.m3u8", fileCode, fileCode)
	masterPlaylist, err := minioClient.GetObject(ctx, user.MinioBucket, masterPlaylistPath, minio.GetObjectOptions{})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	msStat, err := masterPlaylist.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.DataFromReader(http.StatusOK, msStat.Size, "application/vnd.apple.mpegurl", masterPlaylist, nil)
}

func ServeSegment(c *gin.Context) {
	ctx := context.Background()
	fileCode := c.Param("fileCode")
	segmentNum := c.Param("segmentNumber")
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	db := c.MustGet("db").(*gorm.DB)
	minioClient := c.MustGet("minio").(*minio.Client)

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

	segmentPath := fmt.Sprintf("/hls/%s/segment-%s.ts", fileCode, segmentNum)
	log.Println(segmentPath)
	segment, err := minioClient.GetObject(ctx, user.MinioBucket, segmentPath, minio.GetObjectOptions{})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	segmentStat, err := segment.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.DataFromReader(http.StatusOK, segmentStat.Size, "video/MP2T", segment, nil)
}