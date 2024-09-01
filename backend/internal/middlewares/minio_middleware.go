package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
)

func MinIOMiddleware(service services.Service, client *minio.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userClaims := c.MustGet("userClaims").(*utils.UserClaims)

		bucketClient := &models.BucketClient{
			Client: client,
			Bucket: userClaims.Bucket,
			ServiceBucket: userClaims.ServiceBucket,
		}

		service.SetBucketClient(bucketClient)
	}
}