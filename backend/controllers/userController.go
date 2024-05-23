package controllers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt"
	"gorm.io/gorm"
)

func UserCreate(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	validate := validator.New()

	var userBody struct {
		FirstName string `json:"first_name" validate:"required,ascii"`
		LastName  string `json:"last_name" validate:"required,ascii"`
		Email     string `validate:"required,email"`
		Password  string `validate:"required,min=6"`
	}

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

	hashedPassword, err := utils.HashPassword(userBody.Password)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

	user := models.User{FirstName: userBody.FirstName, LastName: userBody.LastName, Email: userBody.Email, Password: hashedPassword}

	err = db.Create(&user).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		log.Println(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": user,
	})
}

func UserLogin(c *gin.Context) {
	db := c.MustGet("db").(*gorm.DB)
	validate := validator.New()

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
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
		},
	}

	accessToken, err := utils.GenerateToken(userClaim)

	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err)
		return
	}

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