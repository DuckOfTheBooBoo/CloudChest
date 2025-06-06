package main

import (
	"fmt"
	"log"
	"os"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/handlers"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/api/routes"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/database"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/database/migrations"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/middlewares"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/services"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	cron "github.com/robfig/cron/v3"
)

func init() {
	utils.LoadEnv()
}

func isDatabaseMigrated(db database.Database) bool {
	var result []string
	rawSQL := fmt.Sprintf("SHOW TABLES FROM %s;", os.Getenv("DB_NAME"))
	db.GetDB().Raw(rawSQL).Scan(&result)
	return len(result) > 0
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	r := gin.Default()

	db, err := database.ConnectToDB()

	if err != nil {
		log.Fatal(err)
		return
	}

	// Check if the database is migrated
	if !isDatabaseMigrated(db) {
		if err := migrations.Migrate(db); err != nil {
			log.Fatal(err)
		}
		log.Println("Database migrated successfully")
	} else {
		log.Println("Database is already migrated")
	}

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 8 << 20

	minioClient, minioErr := database.ConnectToMinIO()

	if minioErr != nil {
		log.Fatal(minioErr)
		return
	}

	// Allow cors
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowHeaders = []string{"*"}
	corsConfig.AllowAllOrigins = true
	corsMiddleware := cors.New(corsConfig)
	r.Use(corsMiddleware)

	// Define root endpoint
	api := r.Group("/api")

	api.Use(middlewares.DBMiddleware(db.GetDB(), minioClient.GetMinioClient()))

	fileService := services.NewFileService(db.GetDB())
	fileHandler := handlers.NewFileHandler(fileService)

	folderService := services.NewFolderService(db.GetDB())
	folderHandler := handlers.NewFolderHandler(folderService)

	userService := services.NewUserService(db.GetDB(), minioClient.GetMinioClient())
	userHandler := handlers.NewUserHandler(userService)

	authService := services.NewAuthService(db.GetDB())
	authHandler := handlers.NewAuthHandler(authService)

	hlsService := services.NewHLSService(db.GetDB(), nil)
	hlsHandler := handlers.NewHLSHandler(hlsService)

	routes.AuthRoutes(api, authHandler)
	routes.TokenRoutes(api)
	routes.UserRoutes(api, userHandler)
	routes.FileRoutes(api, fileHandler, minioClient.GetMinioClient())
	routes.FolderRoutes(api, folderHandler, minioClient.GetMinioClient())
	routes.HLSRoutes(api, hlsHandler, minioClient.GetMinioClient())

	// Schedule revoked tokens ('tokens' table in database) pruning
	c := cron.New()
	cronSpec := "*/15 * * * *" // Run every 15 minutes

	_, err = c.AddFunc(cronSpec, func() {
		utils.PruneRevokedTokens(db.GetDB())
	})

	if err != nil {
		log.Fatal(err)
	}

	c.Start()

	r.Run(":3000")
}
