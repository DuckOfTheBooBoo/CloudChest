package middlewares

import (
	"net/http"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		db := c.MustGet("db").(*gorm.DB)
		tokenString := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")

		userClaims, err := utils.ParseToken(tokenString)
		// Check if token is revoked (user has logged out, marking the token as invalid)
		var token models.Token
		if err := db.First(&token, "token = ?", tokenString); err != nil {
			token = models.Token{}
		}

		if err != nil || (models.Token{}) != token {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set("userClaims", userClaims)
		c.Next()
	}
}	