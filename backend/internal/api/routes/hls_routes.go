package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func HLSRoutes(route *gin.RouterGroup, hlsHandler *handlers.HLSHandler, mc *minio.Client) {
	hlsRouter := route.Group("/hls")
	{
		hlsRouter.GET("/:fileCode/masterPlaylist", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(hlsHandler.HLSService, mc), hlsHandler.ServeMasterPlaylist)
		hlsRouter.GET("/:fileCode/segments/:segmentNumber", middlewares.JWTMiddleware(), middlewares.MinIOMiddleware(hlsHandler.HLSService, mc), hlsHandler.ServeSegment)
	}
}
