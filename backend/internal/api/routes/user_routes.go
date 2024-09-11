package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/gin-gonic/gin"
)

func UserRoutes(route *gin.RouterGroup, userHandler *handlers.UserHandler) {
	user := route.Group("/users") 
	{
		user.POST("/register", userHandler.UserCreate)
		user.PUT("/:userId", middlewares.JWTMiddleware(), func(ctx *gin.Context) {})
		user.DELETE("/:userId", middlewares.JWTMiddleware(), func(ctx *gin.Context) {})
	}
}