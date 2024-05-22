package middlewares

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"strings"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {		
		tokenString := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")

		userClaims, err := utils.ParseToken(tokenString)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}

		c.Set("userClaims", userClaims)
		c.Next()
	}
}	