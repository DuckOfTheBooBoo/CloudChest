package jobs

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
)

func WriteTempFile(file models.File, assetFile io.Reader) string {
	tempPath := fmt.Sprintf("/tmp/%s-file", file.FileCode)

	tempFile, err := os.Create(tempPath)
	if err != nil {
		log.Println(err)
	}

	_, err = io.Copy(tempFile, assetFile)
	if err != nil {
		log.Println(err)
	}

	return tempPath
}
