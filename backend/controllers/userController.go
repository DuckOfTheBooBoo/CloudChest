package controllers

import (
	"log"
	"net/http"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
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
