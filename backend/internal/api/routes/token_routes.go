package routes

import (
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/gin-gonic/gin"
)

// TODO: I'm planning to remove this, we can use interceptors on frontend everytime the request returned 401 unauthenticated error
func TokenRoutes(route *gin.RouterGroup) {
	token := route.Group("/token") 
	{
		token.POST("/check", handlers.CheckTokenValidation)
	}
}