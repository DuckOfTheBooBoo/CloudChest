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
		file.GET("/favorite", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FileFavorites)
		file.GET("/trashcan", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FileTrashCan)
		file.GET("/:fileCode/thumbnail", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FileThumbnail)
		file.GET("/:fileCode/download", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), handlers.FileDownload) // TODO: implement reverse proxy for download
		file.PUT("/:fileID", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FileUpdate)
		file.PATCH("/:fileID", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FilePatch)
		file.DELETE("", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FileDeleteAll)
		file.DELETE("/:fileID", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(fileHandler.FileService, minioClient), fileHandler.FileDelete)
	}
}