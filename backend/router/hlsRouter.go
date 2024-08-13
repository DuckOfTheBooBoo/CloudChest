package router

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/controllers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/middlewares"
	"github.com/gin-gonic/gin"
)

func HLSRoutes(route *gin.RouterGroup) {
	hlsRouter := route.Group("/hls")
	{
		hlsRouter.GET("/:fileCode/masterPlaylist", middlewares.JWTMiddleware(), controllers.ServeMasterPlaylist)
		hlsRouter.GET("/:fileCode/segments/:segmentNumber", middlewares.JWTMiddleware(), controllers.ServeSegment)
	}
}
