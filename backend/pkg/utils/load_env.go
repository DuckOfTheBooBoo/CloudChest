package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func LoadEnv() error {
    if os.Getenv("GO_ENV") == "production" {
        return nil
    }

	dir, err := os.Getwd()
    if err != nil {
        return err
    }

    envPath := filepath.Join(dir, ".env")

    err = godotenv.Load(envPath)
    if err != nil {
        log.Printf("Error loading .env file: %v\n", err)
    }
    return err
}