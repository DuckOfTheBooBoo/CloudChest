package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"
	
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

type UserHandler struct {
	UserService *services.UserService
}

func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func (uf *UserHandler) UserCreate(c *gin.Context) {
	validate := validator.New()
	var userBody models.UserBody
	err := c.BindJSON(&userBody)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No request body (JSON) included.",
		})
		return
	}

	if err := validate.Struct(userBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	user, err := uf.UserService.CreateUser(&userBody)
	switch e := err.(type) {
		case *apperr.InvalidParamError:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": e.Error(),
			})
			return
		case *apperr.ServerError:
			c.Status(http.StatusInternalServerError)
			return
		default:
			break
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
	})
}

func UserLogin(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
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
		Password string `validate:"required,min=6"`
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

	var user models.User

	err = db.First(&user, "email = ?", loginBody.Email).Error

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not found",
		})
		log.Println(err)
		return
	}

	if !utils.CheckPassword(loginBody.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid credentials",
		})
		return
	}

	userClaim := utils.UserClaims{
		ID: user.ID,
		Bucket: user.MinioBucket,
		ServiceBucket: user.MinioServiceBucket,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 60).Unix(),
		},
	}

	accessToken, err := utils.GenerateToken(userClaim)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	c.SetCookie("token", accessToken, 60*60, "/", domain, false, true)

	c.JSON(http.StatusOK, gin.H{
		"token": accessToken,
	})
}

func UserLogout(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	userClaim := c.MustGet("userClaims").(*utils.UserClaims)
	tokenString := strings.ReplaceAll(c.GetHeader("Authorization"), "Bearer ", "")

	token := models.Token{Token: tokenString, ExpirationDate: time.Unix(userClaim.ExpiresAt, 0)}

	err := db.Create(&token).Error

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.Status(http.StatusCreated)
}