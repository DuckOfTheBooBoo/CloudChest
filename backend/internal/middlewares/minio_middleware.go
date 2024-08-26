package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func MinIOMiddleware(client *minio.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("minio", client)
		c.Next()
	}
}