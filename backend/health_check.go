package main

import (
	"log"
	"os"

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

	_, err = database.ConnectToMinIO()

	if err != nil {
		log.Fatal(err)
	}


	var tables = []interface{}{models.User{}, models.Token{}, models.Folder{}, models.File{}, models.Thumbnail{}}
	for _, table := range tables {
		if !db.GetDB().Migrator().HasTable(table) {
			log.Fatal(table)
		}
	}

	os.Exit(0)
}