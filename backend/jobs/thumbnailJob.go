package jobs

import (
	"bytes"
	"context"
	"image/jpeg"
	"log"
	"fmt"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func GenerateThumbnail(ctx context.Context, minioClient *minio.Client, db *gorm.DB, file models.File, userBucket string) {
	assetFile, err := minioClient.GetObject(ctx, userBucket, file.FileCode, minio.GetObjectOptions{})

	if err != nil {
		log.Printf("Error while fetching file: %s -> %s\n", file.FileName, err.Error())
		return
	}
	defer assetFile.Close()

	assetImg, err := imaging.Decode(assetFile, imaging.AutoOrientation(true))
	
	if err != nil {
		log.Printf("Error while decoding file: %s -> %s\n", file.FileName, err.Error())
		return
	}
	
	height := assetImg.Bounds().Dy()
	width := assetImg.Bounds().Dx()

	thumbHeight := float64(150)
	thumbWidth := float64(width) * (thumbHeight / float64(height))

	thumbImg := imaging.Resize(assetImg, int(thumbWidth), int(thumbHeight), imaging.NearestNeighbor)

	// Encode the thumbImg to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbImg, &jpeg.Options{Quality: 85}); err != nil {
		log.Printf("Error while processing thumbnail: %s -> %s\n", file.FileName, err.Error())
		return
	}

	size := int64(buf.Len())
	thumbPath := fmt.Sprintf("/thumb/%s.jpg", file.FileCode)

	_, err = minioClient.PutObject(ctx, userBucket, thumbPath, &buf, size, minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		log.Printf("Error while uploading thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}

	thumbnail := models.Thumbnail{
		FileID: file.ID,
		FilePath: thumbPath,
	}

	if err := db.Create(&thumbnail).Error; err != nil {
		log.Printf("Error while saving thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}
	
	log.Printf("Thumbnail created: %s (%s)\n", thumbPath, file.FileName)
}