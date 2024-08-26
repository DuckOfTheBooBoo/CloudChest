package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/gin-gonic/gin"
)


func TokenRoutes(route *gin.RouterGroup) {
	token := route.Group("/token") 
	{
		token.POST("/check", handlers.CheckTokenValidation)
	}
}