package router

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/controllers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/middlewares"
	"github.com/gin-gonic/gin"
)

func FileRoutes(route *gin.RouterGroup) {
	file := route.Group("/files") 
	{	
		file.GET("", middlewares.JWTMiddleware(), controllers.FileList)
		
		// Get all files inside a folder by code
		file.GET("/:code", middlewares.JWTMiddleware(), controllers.FileList)
		file.GET("/thumbnail/:fileID", middlewares.JWTMiddleware(), controllers.FileThumbnail)
		file.GET("/download/:fileID", middlewares.JWTMiddleware(), controllers.FileDownload)
		file.POST("", middlewares.JWTMiddleware(), controllers.FileUpload)
		file.POST("/:code", middlewares.JWTMiddleware(), controllers.FileUpload)
		file.PUT("/:fileID", middlewares.JWTMiddleware(), controllers.FileUpdate)
		file.DELETE("/:fileID", middlewares.JWTMiddleware(), controllers.FileDelete)
	}
}