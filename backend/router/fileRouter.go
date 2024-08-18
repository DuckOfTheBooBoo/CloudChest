package router

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/controllers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/middlewares"
	"github.com/gin-gonic/gin"
)

func FileRoutes(route *gin.RouterGroup) {
	file := route.Group("/files") 
	{	
		file.GET("", middlewares.JWTMiddleware(), controllers.FileList) // Only for favorites and trashcan
		file.GET("/:fileID/thumbnail", middlewares.JWTMiddleware(), controllers.FileThumbnail)
		file.GET("/:fileID/download", middlewares.JWTMiddleware(), controllers.FileDownload)
		file.PUT("/:fileID", middlewares.JWTMiddleware(), controllers.FileUpdate)
		file.PATCH("/:fileID", middlewares.JWTMiddleware(), controllers.FilePatch)
		file.DELETE("", middlewares.JWTMiddleware(), controllers.FileDeleteAll)
		file.DELETE("/:fileID", middlewares.JWTMiddleware(), controllers.FileDelete)
	}
}