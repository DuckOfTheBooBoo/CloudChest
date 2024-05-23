package main

import (
	"log"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/database"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/middlewares"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/router"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/utils"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	cron "github.com/robfig/cron/v3"
)

func init() {
	utils.LoadEnv()
}

func main() {
	r := gin.Default()

	db, err := database.ConnectToDB()

	if err != nil {
		log.Fatal(err)
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

	api.Use(middlewares.DBMiddleware(db.GetDB()))
	router.TokenRoutes(api)
	router.UserRoutes(api)

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
