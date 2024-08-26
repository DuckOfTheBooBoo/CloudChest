package jobs

import (
	"fmt"
	"os"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
)

func WriteTempFile(file models.File, assetFileBytes []byte) (string, error) {
	tempPath := fmt.Sprintf("/tmp/%s-file", file.FileCode)

	tempFile, err := os.Create(tempPath)
	if err != nil {
		return "", fmt.Errorf("error while creating temp file: %v", err)
	}
	_, err = tempFile.Write(assetFileBytes)
	if err != nil {
		return "", fmt.Errorf("error while writing to temp file: %v", err)
	}

	return tempPath, nil
}
