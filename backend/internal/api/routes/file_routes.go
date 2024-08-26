package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func FileRoutes(route *gin.RouterGroup) {
	file := route.Group("/files") 
	{	
		file.GET("", middlewares.JWTMiddleware(), handlers.FileList) // Only for favorites and trashcan
		file.GET("/:fileID/thumbnail", middlewares.JWTMiddleware(), handlers.FileThumbnail)
		file.GET("/:fileID/download", middlewares.JWTMiddleware(), handlers.FileDownload)
		file.PUT("/:fileID", middlewares.JWTMiddleware(), handlers.FileUpdate)
		file.PATCH("/:fileID", middlewares.JWTMiddleware(), handlers.FilePatch)
		file.DELETE("", middlewares.JWTMiddleware(), handlers.FileDeleteAll)
		file.DELETE("/:fileID", middlewares.JWTMiddleware(), handlers.FileDelete)
	}
}