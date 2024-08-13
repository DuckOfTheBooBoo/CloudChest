package jobs

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/minio/minio-go/v7"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func ProcessHLS(filePath string, ctx context.Context, minioClient *minio.Client, file models.File, userBucket string) {
	tmpDir := "/tmp/"+file.FileCode
	if _, err := os.Stat(tmpDir); err == nil {
		os.RemoveAll(tmpDir)
	}

	log.Println("Creating temporary directory: " + tmpDir)
	os.MkdirAll(tmpDir, 0755)

	log.Println("Processing HLS file: " + file.FileName)
	err := ffmpeg.Input(filePath).
		Output(fmt.Sprintf("%s/%s.m3u8", tmpDir, file.FileCode),
			ffmpeg.KwArgs{
				"codec:":               "copy",
				"hls_time":             10,
				"hls_list_size":        0,
				"f":                    "hls",
				"hls_segment_filename": fmt.Sprintf("%s/segment-%%d.ts", tmpDir),
			}).
		Run()

	if err != nil {
		log.Printf("Error while processing HLS file: %s -> %v", file.FileName, err)
	}

	// Replace all relative path in m3u8 file
	tempMasterPlaylistFilePath := fmt.Sprintf("%s/%s.m3u8", tmpDir, file.FileCode)
	log.Printf("Reading master playlist: %s", tempMasterPlaylistFilePath)
	content, err := os.ReadFile(tempMasterPlaylistFilePath)
	if err != nil {
		log.Printf("Error while reading master playlist: %s -> %v", file.FileName, err)
	}

	text := string(content)

	log.Println("Modifying master playlist")
	modifiedText1 := strings.ReplaceAll(text, "segment", fmt.Sprintf("/api/hls/%s/segments/segment", file.FileCode))
	modifiedText2 := strings.ReplaceAll(modifiedText1, "segment-", "")
	modifiedText := strings.ReplaceAll(modifiedText2, ".ts", "")

	log.Println("Saving modified master playlist")
	err = os.WriteFile(fmt.Sprintf("%s/%s.m3u8", tmpDir, file.FileCode), []byte(modifiedText), 0644)
	if err != nil {
		log.Printf("Error while saving modified master playlist: %s -> %v\n", file.FileName, err)
	}

	log.Println("Reading temporary files")
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		log.Println(err)
	}

	log.Println("Uploading HLS files")
	for _, hlsFile := range files {
		if !hlsFile.IsDir() {
			filePath := fmt.Sprintf("%s/%s", tmpDir, hlsFile.Name())

			fileData, err := os.Open(filePath)
			if err != nil {
				log.Printf("Error while reading file: %s -> %v\n", filePath, err)
				return
			}

			log.Println("Reading file info")
			fileInfo, err := fileData.Stat()
			if err != nil {
				log.Printf("Error while reading file info: %s -> %v\n", filePath, err)
				return
			}
			fileSize := fileInfo.Size()
			contentType := ""
			if strings.HasSuffix(hlsFile.Name(), "m3u8") {
				contentType = "application/vnd.apple.mpegurl"
			} else {
				contentType = "video/MP2T"
			}
			
			// Upload hlsFile to MinIO
			hlsFilePath := fmt.Sprintf("hls/%s/%s", file.FileCode, hlsFile.Name())

			log.Println("Uploading HLS file: " + hlsFilePath + " to " + userBucket)
			_, err = minioClient.PutObject(context.Background(), userBucket, hlsFilePath, fileData, fileSize, minio.PutObjectOptions{ContentType: contentType})
			if err != nil {
				log.Printf("Error while uploading file: %s -> %v\n", filePath, err)
				return
			}
		}
	}

	log.Println("Created HLS playlist: " + file.FileCode)
}