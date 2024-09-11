package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

func DBMiddleware(db *gorm.DB, minioClient *minio.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Set("minio", minioClient)
		c.Next()
	}
}