package middlewares

import (
	"net/http"
	"strings"

	_ "github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	_ "gorm.io/gorm"
)

func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// db := c.MustGet("db").(*gorm.DB)
		tokenString := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")
		tokenCookie, err := c.Cookie("token")

		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		if tokenCookie != "" {
			tokenString = tokenCookie
		}

		userClaims, err := utils.ParseToken(tokenString)
		// Check if token is revoked (user has logged out, marking the token as invalid)
		// var token models.Token
		// if err := db.First(&token, "token = ?", tokenString); err != nil {
		// 	token = models.Token{}
		// }

		// if err != nil || (models.Token{}) != token {
		if err != nil {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Set("userClaims", userClaims)
		c.Next()
	}
}	