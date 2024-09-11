package middlewares

import (
	"context"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func MinIOMiddleware(service services.Service, client *minio.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userClaims := c.MustGet("userClaims").(*utils.UserClaims)

		bucketClient := &models.BucketClient{
			Context: context.Background(),
			Client: client,
			Bucket: userClaims.Bucket,
			ServiceBucket: userClaims.ServiceBucket,
		}

		service.SetBucketClient(bucketClient)
	}
}