package services

import (
	"context"
	"log"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type HLSService struct {
	DB *gorm.DB
	BucketClient *models.BucketClient
}

func (hs *HLSService) SetDB(db *gorm.DB) {
	hs.DB = db
}

func (hs *HLSService) SetBucketClient(bc *models.BucketClient) {
	hs.BucketClient = bc
}

func NewHLSService(db *gorm.DB, bc *models.BucketClient) *HLSService {
	return &HLSService{
		DB: db,
		BucketClient: bc,
	}
}


func (hs *HLSService) DeleteHLSFiles(file *models.File) error {
	ctx := context.Background()

	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range hs.BucketClient.Client.ListObjects(ctx, hs.BucketClient.ServiceBucket, minio.ListObjectsOptions{
			Prefix:    "hls/" + file.FileCode,
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Println("Error listing HLS files: ", object.Err)
				continue
			}
			if strings.HasSuffix(object.Key, ".m3u8") || strings.HasSuffix(object.Key, ".ts") {
				objectsCh <- object
			}
		}
	}()

	for rErr := range hs.BucketClient.Client.RemoveObjects(ctx, hs.BucketClient.ServiceBucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if rErr.Err != nil {
			return &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err: rErr.Err,
				},
			}
		}
	}
	return nil
}