package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func FolderRoutes(route *gin.RouterGroup, folderHandler *handlers.FolderHandler) {
	folder := route.Group("/folders") 
	{	
		folder.GET("", middlewares.JWTMiddleware(), folderHandler.FolderList)
		folder.GET("/:code", middlewares.JWTMiddleware(), folderHandler.FolderList)
		folder.PATCH("/:code", middlewares.JWTMiddleware(), folderHandler.FolderPatch)
		folder.GET("/:code/files", middlewares.JWTMiddleware(), handlers.FolderContents)
		folder.GET("/:code/folders", middlewares.JWTMiddleware(), folderHandler.FolderList)
		folder.POST("/:code/files", middlewares.JWTMiddleware(), handlers.FolderContentsCreate)
		folder.POST("/:code/folders", middlewares.JWTMiddleware(), handlers.FolderCreate)
		folder.DELETE("/:code", middlewares.JWTMiddleware(), handlers.FolderDelete)
	}
}