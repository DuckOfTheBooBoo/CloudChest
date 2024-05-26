package router

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/controllers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/middlewares"
	"github.com/gin-gonic/gin"
)

func FileRoutes(route *gin.RouterGroup) {
	file := route.Group("/files") 
	{	
		file.GET("/:path", middlewares.JWTMiddleware(), controllers.FileList)
		file.POST("", middlewares.JWTMiddleware(), controllers.FileUpload)
		file.POST("/path", middlewares.JWTMiddleware(), controllers.FileNewPath)
		file.DELETE("/:fileID", middlewares.JWTMiddleware(), controllers.FileDelete)
	}
}