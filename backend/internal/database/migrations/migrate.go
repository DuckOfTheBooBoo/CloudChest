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
	log.Println("(Migrate) Connected to DB")

	if err != nil {
		log.Fatal(err)
	}

	gormDB := db.GetDB()

	log.Println("(Migrate) Migrating...")
	migErr := gormDB.AutoMigrate(&models.User{}, &models.Token{}, &models.Folder{}, &models.File{}, &models.Thumbnail{})

	if migErr != nil {
		log.Println("(Migrate) Migration Failed")
		log.Fatal(migErr)
	}
	log.Println("(Migrate) Migration Successful")
}
