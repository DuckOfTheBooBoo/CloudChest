package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"math/rand"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/internal/models"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/pkg/apperr"
	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gorm.io/gorm"
)

type ThumbnailService struct {
	DB *gorm.DB
	BucketClient *models.BucketClient
}

func (ts *ThumbnailService) SetDB(db *gorm.DB) {
	ts.DB = db
}

func (ts *ThumbnailService) SetBucketClient(bc *models.BucketClient) {
	ts.BucketClient = bc
}

func NewThumbnailService(db *gorm.DB, bc *models.BucketClient) *ThumbnailService {
	return &ThumbnailService{
		DB: db,
		BucketClient: bc,
	}
}


type Stream struct {
	Index              int         `json:"index"`
	CodecName          string      `json:"codec_name"`
	CodecLongName      string      `json:"codec_long_name"`
	Profile            string      `json:"profile"`
	CodecType          string      `json:"codec_type"`
	CodecTagString     string      `json:"codec_tag_string"`
	CodecTag           string      `json:"codec_tag"`
	Width              int         `json:"width"`
	Height             int         `json:"height"`
	CodedWidth         int         `json:"coded_width"`
	CodedHeight        int         `json:"coded_height"`
	ClosedCaptions     int         `json:"closed_captions"`
	FilmGrain          int         `json:"film_grain"`
	HasBFrames         int         `json:"has_b_frames"`
	SampleAspectRatio  string      `json:"sample_aspect_ratio"`
	DisplayAspectRatio string      `json:"display_aspect_ratio"`
	PixFmt             string      `json:"pix_fmt"`
	Level              int         `json:"level"`
	ColorRange         string      `json:"color_range"`
	ColorSpace         string      `json:"color_space"`
	ColorTransfer      string      `json:"color_transfer"`
	ColorPrimaries     string      `json:"color_primaries"`
	ChromaLocation     string      `json:"chroma_location"`
	FieldOrder         string      `json:"field_order"`
	Refs               int         `json:"refs"`
	IsAvc              string      `json:"is_avc"`
	NalLengthSize      string      `json:"nal_length_size"`
	ID                 string      `json:"id"`
	RFrameRate         string      `json:"r_frame_rate"`
	AvgFrameRate       string      `json:"avg_frame_rate"`
	TimeBase           string      `json:"time_base"`
	StartPts           int         `json:"start_pts"`
	StartTime          string      `json:"start_time"`
	DurationTs         int         `json:"duration_ts"`
	Duration           string      `json:"duration"`
	BitRate            string      `json:"bit_rate"`
	BitsPerRawSample   string      `json:"bits_per_raw_sample"`
	NbFrames           string      `json:"nb_frames"`
	ExtradataSize      int         `json:"extradata_size"`
	Disposition        Disposition `json:"disposition"`
	Tags               Tags        `json:"tags"`
	SampleFmt          string      `json:"sample_fmt"`
	SampleRate         string      `json:"sample_rate"`
	Channels           int         `json:"channels"`
	ChannelLayout      string      `json:"channel_layout"`
	BitsPerSample      int         `json:"bits_per_sample"`
	InitialPadding     int         `json:"initial_padding"`
}

type Disposition struct {
	Default         int `json:"default"`
	Dub             int `json:"dub"`
	Original        int `json:"original"`
	Comment         int `json:"comment"`
	Lyrics          int `json:"lyrics"`
	Karaoke         int `json:"karaoke"`
	Forced          int `json:"forced"`
	HearingImpaired int `json:"hearing_impaired"`
	VisualImpaired  int `json:"visual_impaired"`
	CleanEffects    int `json:"clean_effects"`
	AttachedPic     int `json:"attached_pic"`
	TimedThumbnails int `json:"timed_thumbnails"`
	Captions        int `json:"captions"`
	Descriptions    int `json:"descriptions"`
	Metadata        int `json:"metadata"`
	Dependent       int `json:"dependent"`
	StillImage      int `json:"still_image"`
}
type Tags struct {
	Language    string `json:"language"`
	HandlerName string `json:"handler_name"`
	VendorID    string `json:"vendor_id"`
}
type Format struct {
	Filename       string     `json:"filename"`
	NbStreams      int        `json:"nb_streams"`
	NbPrograms     int        `json:"nb_programs"`
	FormatName     string     `json:"format_name"`
	FormatLongName string     `json:"format_long_name"`
	StartTime      string     `json:"start_time"`
	Duration       string     `json:"duration"`
	Size           string     `json:"size"`
	BitRate        string     `json:"bit_rate"`
	ProbeScore     int        `json:"probe_score"`
	Tags           FormatTags `json:"tags"`
}

type FormatTags struct {
	MajorBrand       string `json:"major_brand"`
	MinorVersion     string `json:"minor_version"`
	CompatibleBrands string `json:"compatible_brands"`
	Encoder          string `json:"encoder"`
}

type ProbeResult struct {
	Streams []Stream `json:"streams"`
	Format  Format   `json:"format"`
}

func getRandomFrameNum(start int, end int) int {
	return rand.Intn(end-start) + start // Generate random number within range
}

func processImage(file *models.File, filePath string) (bytes.Buffer, error) {
	log.Println("Processing image thumbnail for: " + filePath)
	assetFile, err := os.Open(filePath)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error while opening image file %s: %v", file.FileName, err)
	}
	defer assetFile.Close()

	assetImg, err := imaging.Decode(assetFile, imaging.AutoOrientation(true))
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("error while decoding image file %s: %v", file.FileName, err)
	}

	height := assetImg.Bounds().Dy()
	width := assetImg.Bounds().Dx()

	thumbHeight := float64(150)
	thumbWidth := float64(width) * (thumbHeight / float64(height))

	thumbImg := imaging.Resize(assetImg, int(thumbWidth), int(thumbHeight), imaging.NearestNeighbor)

	// Encode the thumbImg to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbImg, &jpeg.Options{Quality: 85}); err != nil {
		return bytes.Buffer{}, fmt.Errorf("error while processing thumbnail: %s -> %v", file.FileName, err)
	}

	return buf, nil	
}

func processVideo(filePath string, debug bool) (bytes.Buffer, error) {
	log.Println("Processing video thumbnail for: " + filePath)
	str, err := ffmpeg.Probe(filePath)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to probe video file: %v", err)
	}

	metadataJSONBytes := []byte(str)

	var probeResult ProbeResult
	err = json.Unmarshal(metadataJSONBytes, &probeResult)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to parse ffprobe output: %v", err)
	}

	// Get the number of video's frames
	var stream Stream
	for _, s := range probeResult.Streams {
		if s.CodecType == "video" {
			stream = s
			break
		}
	}

	if stream == (Stream{}) {
		return bytes.Buffer{}, fmt.Errorf("failed to find video stream")
	}

	numFrames := stream.NbFrames
	numFramesInt := 0

	if numFrames == "" {
		// Get the frame rate
		rFrames := strings.Split(stream.RFrameRate, "/")

		if len(rFrames) != 2 {
			return bytes.Buffer{}, fmt.Errorf("failed to parse frame rate")
		}

		num1, err1 := strconv.Atoi(rFrames[0])
		num2, err2 := strconv.Atoi(rFrames[1])
		if err1 != nil || err2 != nil {
			return bytes.Buffer{}, fmt.Errorf("failed to parse string to int: %v", err)
		}

		frameRate := math.Floor(float64(num1) / float64(num2))

		// Calculate the number of frames
		duration, err := strconv.ParseFloat(stream.Duration, 64)
		if err != nil {
			return bytes.Buffer{}, fmt.Errorf("failed to parse duration to float: %v", err)
		}

		numFramesInt = int(duration * frameRate)
	} else {
		numFramesInt, err = strconv.Atoi(numFrames)
		if err != nil {
			return bytes.Buffer{}, fmt.Errorf("failed to parse string to int: %v", err)
		}
	}

	frameThumbnail := getRandomFrameNum(0, numFramesInt)
	outBuf := new(bytes.Buffer)
	
	var errorOutput io.Writer
	if debug {
        errorOutput = os.Stderr // Show output in debug mode
    } else {
        errorOutput = io.Discard // Hide output in normal mode
    }

	err = ffmpeg.Input(filePath).
		Filter("select", ffmpeg.Args{fmt.Sprintf(`eq(n\,%d)`, frameThumbnail)}).
		Filter("scale", ffmpeg.Args{"-1", "150"}).
		Output("-", ffmpeg.KwArgs{
			"vframes": "1","f": "image2pipe"}).
		WithOutput(outBuf).
		OverWriteOutput().  // Redirect FFmpeg logs to standard output and error.                // Add this if you want to overwrite the output file if it exists
		WithErrorOutput(errorOutput).                  // Ensure that error messages go to standard output
		Run()
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("failed to generate video thumbnail: %v", err)
	}

	return *outBuf, nil
}

func (ts *ThumbnailService) GenerateThumbnail(filePath string, file *models.File) {
	var thumbnailBuf bytes.Buffer
	var err error

	if filePath == "" {
		log.Println("File path is empty")
		return
	}

	if strings.HasPrefix(file.FileType, "image/") {
		thumbnailBuf, err = processImage(file, filePath)
		if err != nil {
			log.Printf("Error while generating image thumbnail: %s -> %v\n", file.FileName, err)
			return
		}
	} else if strings.HasPrefix(file.FileType, "video/") {
		thumbnailBuf, err = processVideo(filePath, false)
		if err != nil {
			log.Printf("Error while generating video thumbnail: %s -> %v\n", file.FileName, err)
			return
		}
	}

	size := int64(thumbnailBuf.Len())
	thumbPath := fmt.Sprintf("/thumb/%s.jpg", file.FileCode)

	_, err = ts.BucketClient.PutServiceObject(thumbPath, &thumbnailBuf, size, minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		log.Printf("Error while uploading thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}

	thumbnail := models.Thumbnail{
		FileID:   file.ID,
		FilePath: thumbPath,
	}

	if err := ts.DB.Create(&thumbnail).Error; err != nil {
		log.Printf("Error while saving thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}

	log.Printf("Thumbnail created: %s (%s)\n", thumbPath, file.FileName)
}

func (ts *ThumbnailService) DeleteThumbnail(thumbnail *models.Thumbnail) error {
	if err := ts.DB.Unscoped().Delete(thumbnail).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "thumbnail not found",
				},
			}
		}

		return &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}	

	if err := ts.BucketClient.RemoveObject(thumbnail.FilePath, minio.RemoveObjectOptions{}); err != nil {
		return &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	return nil
}

func (ts *ThumbnailService) GetThumbnail(fileCode string, userID uint, isDeleted bool) (*minio.Object, error) {
	var file models.File

	query := ts.DB.Model(&models.File{}).Where("file_code = ? AND user_id = ?", fileCode, userID)

	if isDeleted {
		query = query.Unscoped()
	}

	if err := query.Preload("Thumbnail").First(&file).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &apperr.NotFoundError{
				BaseError: &apperr.BaseError{
					Message: "file not found",
					Err: err,
				},
			}
		}

		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}

	if file.Thumbnail == nil {
		if strings.HasPrefix(file.FileType, "image/") || strings.HasPrefix(file.FileType, "video/") {
			return nil, &apperr.ResourceNotReadyError{
				BaseError: &apperr.BaseError{
					Message: "file's thumbnail is being processed",
				},
			}
		} else {
			return nil, &apperr.InvalidParamError{
				BaseError: &apperr.BaseError{
					Message: "file is not an image or a video",
				},
			}
		}
	}

	// Close at handler
	thumbnail, err := ts.BucketClient.GetServiceObject(file.Thumbnail.FilePath, minio.GetObjectOptions{})
	if err != nil {
		return nil, &apperr.ServerError{
			BaseError: &apperr.BaseError{
				Message: "Internal server error ocurred",
				Err: err,
			},
		}
	}
	
	return thumbnail, nil
}