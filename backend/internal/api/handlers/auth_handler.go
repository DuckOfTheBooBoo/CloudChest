package handlers

import (
	"net/http"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	AuthService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		AuthService: authService,
	}
}

func (ah *AuthHandler) UserLogin(c *gin.Context) {
	validate := validator.New()
	domain := c.Query("referer")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Referer domain is missing",
		})
		return
	}

	var loginBody struct {
		Email    string `validate:"required,email"`
		Password string `validate:"required"`
	}

	err := c.BindJSON(&loginBody)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No request body (JSON) included.",
		})
		return
	}

	if err := validate.Struct(loginBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	accessToken, err := ah.AuthService.Login(loginBody.Email, loginBody.Password)
	if err != nil {
		switch e := err.(type) {
			case *apperr.NotFoundError:
				c.JSON(http.StatusNotFound, gin.H{
					"error": e.Error(),
				})
				return
			case *apperr.InvalidCredentialsError:
				c.Status(http.StatusUnauthorized)
				return
			case *apperr.ServerError:
				c.Status(http.StatusInternalServerError)
				return
		}
	}

	c.SetCookie("token", accessToken, 60*60, "/", c.Request.Host, false, true)

	c.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}

func (ah *AuthHandler) UserLogout(c *gin.Context) {
	c.SetCookie("token", "", -1, "/", c.Request.Host, false, true)
	c.Status(http.StatusOK)
}