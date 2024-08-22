package router

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/controllers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/middlewares"
	"github.com/gin-gonic/gin"
)

func FolderRoutes(route *gin.RouterGroup) {
	folder := route.Group("/folders") 
	{	
		folder.GET("", middlewares.JWTMiddleware(), controllers.FolderList)
		folder.GET("/:code", middlewares.JWTMiddleware(), controllers.FolderList)
		folder.PATCH("/:code", middlewares.JWTMiddleware(), controllers.FolderPatch)
		folder.GET("/:code/files", middlewares.JWTMiddleware(), controllers.FolderContents)
		folder.GET("/:code/folders", middlewares.JWTMiddleware(), controllers.FolderList)
		folder.POST("/:code/files", middlewares.JWTMiddleware(), controllers.FolderContentsCreate)
		folder.POST("/:code/folders", middlewares.JWTMiddleware(), controllers.FolderCreate)
	}
}