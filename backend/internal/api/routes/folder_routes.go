package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func FolderRoutes(route *gin.RouterGroup, folderHandler *handlers.FolderHandler, mc *minio.Client) {
	folder := route.Group("/folders") 
	{	
		folder.GET("", middlewares.JWTMiddleware(), folderHandler.FolderList)
		folder.GET("/:code", middlewares.JWTMiddleware(), folderHandler.FolderList)
		// TODO: implement /favorite and /trashcan
		folder.PATCH("/:code", middlewares.JWTMiddleware(), folderHandler.FolderPatch)
		folder.GET("/:code/files", middlewares.JWTMiddleware(), folderHandler.FolderContents)
		folder.GET("/:code/folders", middlewares.JWTMiddleware(), folderHandler.FolderList)
		folder.POST("/:code/files", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(folderHandler.FolderService, mc), folderHandler.FolderContentsCreate)
		folder.POST("/:code/folders", middlewares.JWTMiddleware(), folderHandler.FolderCreate)
		folder.DELETE("/:code", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(folderHandler.FolderService, mc), folderHandler.FolderDelete)
	}
}