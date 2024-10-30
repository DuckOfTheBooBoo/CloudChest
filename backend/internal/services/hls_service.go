package services

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/minio/minio-go/v7"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gorm.io/gorm"
)

type HLSService struct {
	DB *gorm.DB
	BucketClient *models.BucketClient
}

func (hs *HLSService) SetDB(db *gorm.DB) {
	hs.DB = db
}

func (hs *HLSService) SetBucketClient(bc *models.BucketClient) {
	hs.BucketClient = bc
}

func NewHLSService(db *gorm.DB, bc *models.BucketClient) *HLSService {
	return &HLSService{
		DB: db,
		BucketClient: bc,
	}
}

// GetMasterPlaylist fetches the master playlist of a HLS file given its file code.
//
// It returns the master playlist object, its size, and an error if any. If the file
// does not exist, it returns a NotFoundError. If there was an internal server error,
// it returns a ServerError.
func (hs *HLSService) GetMasterPlaylist(fileCode string) (*minio.Object, *int64, error) {
	masterPlaylistPath := fmt.Sprintf("/hls/%s/%s.m3u8", fileCode, fileCode)
	masterPlaylist, err := hs.BucketClient.GetServiceObject(masterPlaylistPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to fetch master playlist",
				Err: err,
			},
		}
	}

	if masterPlaylist == nil {
		return nil, nil, &apperr.NotFoundError{
			BaseError: &apperr.BaseError{
				Message: "Master playlist not found",
				Err: err,
			},
		}
	}

	msStat, err := masterPlaylist.Stat()
	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to read master playlist",
				Err: err,
			},
		}
	}

	return masterPlaylist, &msStat.Size, nil
}

// GetSegment fetches a segment of a HLS file given its file code and segment number.
//
// It returns the segment object, its size, and an error if any. If the file
// does not exist, it returns a NotFoundError. If there was an internal server error,
// it returns a ServerError.
func (hs *HLSService) GetSegment(fileCode, segNum string) (*minio.Object, *int64, error) {
	segmentPath := fmt.Sprintf("/hls/%s/segment-%s.ts", fileCode, segNum)
	// log.Println(segmentPath)
	segment, err := hs.BucketClient.GetServiceObject(segmentPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to fetch segment",
				Err: err,
			},
		}
	}

	if segment == nil {
		return nil, nil, &apperr.NotFoundError{
			BaseError: &apperr.BaseError{
				Message: "Segment not found",
				Err: err,
			},
		}
	}

	segmentStat, err := segment.Stat()
	if err != nil {
		return nil, nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Failed to read segment",
				Err: err,
			},
		}
	}

	return segment, &segmentStat.Size, nil
}

func (hs *HLSService) ProcessHLS(filePath string, file *models.File) {
	tmpDir := "/tmp/"+file.FileCode
	if _, err := os.Stat(tmpDir); err == nil {
		os.RemoveAll(tmpDir)
	}

	// log.Println("Creating temporary directory: " + tmpDir)
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
		return
	}

	// Replace all relative path in m3u8 file
	tempMasterPlaylistFilePath := fmt.Sprintf("%s/%s.m3u8", tmpDir, file.FileCode)
	// log.Printf("Reading master playlist: %s", tempMasterPlaylistFilePath)
	content, err := os.ReadFile(tempMasterPlaylistFilePath)
	if err != nil {
		log.Printf("Error while reading master playlist: %s -> %v", file.FileName, err)
		return
	}

	text := string(content)

	// log.Println("Modifying master playlist")
	modifiedText1 := strings.ReplaceAll(text, "segment", fmt.Sprintf("/api/hls/%s/segments/segment", file.FileCode))
	modifiedText2 := strings.ReplaceAll(modifiedText1, "segment-", "")
	modifiedText := strings.ReplaceAll(modifiedText2, ".ts", "")

	// log.Println("Saving modified master playlist")
	err = os.WriteFile(fmt.Sprintf("%s/%s.m3u8", tmpDir, file.FileCode), []byte(modifiedText), 0644)
	if err != nil {
		log.Printf("Error while saving modified master playlist: %s -> %v\n", file.FileName, err)
		return
	}

	// log.Println("Reading temporary files")
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		log.Println(err)
		return
	}

	// log.Println("Uploading HLS files")
	for _, hlsFile := range files {
		if !hlsFile.IsDir() {
			filePath := fmt.Sprintf("%s/%s", tmpDir, hlsFile.Name())

			fileData, err := os.Open(filePath)
			if err != nil {
				log.Printf("Error while reading file: %s -> %v\n", filePath, err)
				return
			}

			// log.Println("Reading file info")
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

			// log.Println("Uploading HLS file: " + hlsFilePath + " to " + userBucket)
			_, err = hs.BucketClient.PutServiceObject(hlsFilePath, fileData, fileSize, minio.PutObjectOptions{ContentType: contentType})
			if err != nil {
				log.Printf("Error while uploading file: %s -> %v\n", filePath, err)
				return
			}
		}
	}

	if err := hs.DB.Model(&file).Update("is_previewable", true).Error; err != nil {
		log.Printf("Error while updating asset file in database: %v", err)
		return
	}

	log.Println("Created HLS playlist: " + file.FileCode)
}

func (hs *HLSService) DeleteHLSFiles(file *models.File) error {
	ctx := context.Background()

	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range hs.BucketClient.Client.ListObjects(ctx, hs.BucketClient.ServiceBucket, minio.ListObjectsOptions{
			Prefix:    "hls/" + file.FileCode,
			Recursive: true,
		}) {
			if object.Err != nil {
				log.Println("Error listing HLS files: ", object.Err)
				continue
			}
			if strings.HasSuffix(object.Key, ".m3u8") || strings.HasSuffix(object.Key, ".ts") {
				objectsCh <- object
			}
		}
	}()

	for rErr := range hs.BucketClient.Client.RemoveObjects(ctx, hs.BucketClient.ServiceBucket, objectsCh, minio.RemoveObjectsOptions{}) {
		if rErr.Err != nil {
			return &apperr.ServerError{
				BaseError: &apperr.BaseError{
					Message: "Internal server error ocurred",
					Err: rErr.Err,
				},
			}
		}
	}
	return nil
}