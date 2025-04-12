package migrations

import (
	"log"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/database"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
)

func init() {
	utils.LoadEnv()
}

func Migrate(db database.Database) error {  
	gormDB := db.GetDB()

	log.Println("(Migrate) Migrating...")
	migErr := gormDB.AutoMigrate(&models.User{}, &models.Token{}, &models.Folder{}, &models.File{}, &models.Thumbnail{})

  	return migErr
}
