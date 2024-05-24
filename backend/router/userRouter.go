package router

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/controllers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/middlewares"
	"github.com/gin-gonic/gin"
)

func UserRoutes(route *gin.RouterGroup) {
	user := route.Group("/users") 
	{
		user.POST("/register", controllers.UserCreate)
		user.POST("/login", controllers.UserLogin)
		user.POST("/logout", middlewares.JWTMiddleware(), controllers.UserLogout)
		user.PUT("/:userId", middlewares.JWTMiddleware(), func(ctx *gin.Context) {})
		user.DELETE("/:userId", middlewares.JWTMiddleware(), func(ctx *gin.Context) {})
	}
}