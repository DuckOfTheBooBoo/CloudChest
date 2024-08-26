package main

import (
	"log"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/database"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
)

func init() {
	utils.LoadEnv()
}

func main() {
	db, err := database.ConnectToDB()

	if err != nil {
		log.Fatal(err)
	}

	gormDB := db.GetDB()

	migErr := gormDB.AutoMigrate(&models.User{}, &models.Token{}, &models.Folder{}, &models.File{}, &models.Thumbnail{})

	if migErr != nil {
		log.Fatal(migErr)
	}
}
