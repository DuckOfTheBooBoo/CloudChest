package router

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/controllers"
	"github.com/gin-gonic/gin"
)


func TokenRoutes(route *gin.RouterGroup) {
	token := route.Group("/token") 
	{
		token.POST("/check", controllers.CheckTokenValidation)
	}
}