package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func FileRoutes(route *gin.RouterGroup, fileHandler *handlers.FileHandler, minioClient *minio.Client) {
	file := route.Group("/files") 
	{	
		file.GET("", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FileList) // Only for favorites and trashcan
		file.GET("/:fileID/thumbnail", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), handlers.FileThumbnail)
		file.GET("/:fileID/download", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), handlers.FileDownload)
		file.PUT("/:fileID", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), handlers.FileUpdate)
		file.PATCH("/:fileID", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), handlers.FilePatch)
		file.DELETE("", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), handlers.FileDeleteAll)
		file.DELETE("/:fileID", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), handlers.FileDelete)
	}
}