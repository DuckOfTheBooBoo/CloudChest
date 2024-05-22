package main

import (
	"log"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/database"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
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

	migErr := gormDB.AutoMigrate(&models.User{})

	if migErr != nil {
		log.Fatal(migErr)
	}
}
