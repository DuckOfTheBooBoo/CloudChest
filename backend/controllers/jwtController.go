package controllers

import (
	"net/http"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
)


func CheckTokenValidation(c *gin.Context) {
	token := c.GetHeader("Authorization")
	token = strings.ReplaceAll(token, "Bearer ", "")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No token provided.",
		})
		return
	}
	
	_, err := utils.ParseToken(token)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Status(http.StatusOK)
}