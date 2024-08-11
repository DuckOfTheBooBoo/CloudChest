package jobs

import (
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"log"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gorm.io/gorm"
)

func processImage(file models.File, assetFile *minio.Object, userBucket string) (bytes.Buffer, error) {
	assetImg, err := imaging.Decode(assetFile, imaging.AutoOrientation(true))

	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("Error while decoding image file %s: %v", file.FileName, err)
	}

	height := assetImg.Bounds().Dy()
	width := assetImg.Bounds().Dx()

	thumbHeight := float64(150)
	thumbWidth := float64(width) * (thumbHeight / float64(height))

	thumbImg := imaging.Resize(assetImg, int(thumbWidth), int(thumbHeight), imaging.NearestNeighbor)

	// Encode the thumbImg to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbImg, &jpeg.Options{Quality: 85}); err != nil {
		return bytes.Buffer{}, fmt.Errorf("Error while processing thumbnail: %s -> %v\n", file.FileName, err)
	}

	return buf, nil

	
}

func processVideo(minioClient *minio.Client, db *gorm.DB, file models.File, assetFile *minio.Object, userBucket string) {
	
}

func GenerateThumbnail(ctx context.Context, minioClient *minio.Client, db *gorm.DB, file models.File, userBucket string) {
	assetFile, err := minioClient.GetObject(ctx, userBucket, file.FileCode, minio.GetObjectOptions{})

	if err != nil {
		log.Printf("Error while fetching file: %s -> %s\n", file.FileName, err.Error())
		return
	}
	defer assetFile.Close()

	var thumbnailBuf bytes.Buffer

	if strings.HasPrefix(file.FileType, "image/") {
		thumbnailBuf, err = processImage(file, assetFile, userBucket)
		if err != nil {
			log.Printf("Error while generating image thumbnail: %s -> %v\n", file.FileName, err)
			return
		}
	} else if strings.HasPrefix(file.FileType, "video/") {
		processVideo(minioClient, db, file, assetFile, userBucket)
	}

	size := int64(thumbnailBuf.Len())
	thumbPath := fmt.Sprintf("/thumb/%s.jpg", file.FileCode)

	_, err = minioClient.PutObject(ctx, userBucket, thumbPath, &thumbnailBuf, size, minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		log.Printf("Error while uploading thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}

	thumbnail := models.Thumbnail{
		FileID:   file.ID,
		FilePath: thumbPath,
	}

	if err := db.Create(&thumbnail).Error; err != nil {
		log.Printf("Error while saving thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}

	log.Printf("Thumbnail created: %s (%s)\n", thumbPath, file.FileName)
}
