package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(route *gin.RouterGroup, authHandler *handlers.AuthHandler) {
	auth := route.Group("/auth") 
	{
		auth.POST("/login", authHandler.UserLogin)
	}
}