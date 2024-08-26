package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func HLSRoutes(route *gin.RouterGroup) {
	hlsRouter := route.Group("/hls")
	{
		hlsRouter.GET("/:fileCode/masterPlaylist", middlewares.JWTMiddleware(), handlers.ServeMasterPlaylist)
		hlsRouter.GET("/:fileCode/segments/:segmentNumber", middlewares.JWTMiddleware(), handlers.ServeSegment)
	}
}
