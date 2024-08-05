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
		folder.POST("", middlewares.JWTMiddleware(), controllers.FolderCreate)
		folder.POST("/:code", middlewares.JWTMiddleware(), controllers.FolderCreate)
	}
}