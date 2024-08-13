package jobs

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
	"github.com/disintegration/imaging"
	"github.com/minio/minio-go/v7"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"gorm.io/gorm"
)


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

func check(msg string, err error) {
	if err != nil {
		log.Printf(msg, err.Error())
	}
}

func getRandomFrameNum(start int, end int) int {
	return rand.Intn(end-start) + start // Generate random number within range
}

func processImage(file models.File, filePath string) (bytes.Buffer, error) {
	log.Println("Processing image thumbnail for: " + filePath)
	assetFile, err := os.Open(filePath)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("Error while opening image file %s: %v", file.FileName, err)
	}
	defer assetFile.Close()

	assetImg, err := imaging.Decode(assetFile, imaging.AutoOrientation(true))
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("Error while decoding image file %s: %v", file.FileName, err)
	}

	height := assetImg.Bounds().Dy()
	width := assetImg.Bounds().Dx()

	thumbHeight := float64(150)
	thumbWidth := float64(width) * (thumbHeight / float64(height))

	thumbImg := imaging.Resize(assetImg, int(thumbWidth), int(thumbHeight), imaging.NearestNeighbor)

	// Encode the thumbImg to JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, thumbImg, &jpeg.Options{Quality: 85}); err != nil {
		return bytes.Buffer{}, fmt.Errorf("Error while processing thumbnail: %s -> %v\n", file.FileName, err)
	}

	return buf, nil	
}

func processVideo(filePath string, debug bool) (bytes.Buffer, error) {
	log.Println("Processing video thumbnail for: " + filePath)
	str, err := ffmpeg.Probe(filePath)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("Failed to probe video file: %v", err)
	}

	metadataJSONBytes := []byte(str)

	var probeResult ProbeResult
	err = json.Unmarshal(metadataJSONBytes, &probeResult)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("Failed to parse ffprobe output: %v", err)
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
		return bytes.Buffer{}, fmt.Errorf("Failed to find video stream")
	}

	numFrames := stream.NbFrames
	numFramesInt, err := strconv.Atoi(numFrames)
	if err != nil {
		return bytes.Buffer{}, fmt.Errorf("Failed to parse string to int: %v", err)
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
		return bytes.Buffer{}, fmt.Errorf("Failed to generate video thumbnail: %v", err)
	}

	return *outBuf, nil
}

func GenerateThumbnail(ctx context.Context, filePath string, minioClient *minio.Client, db *gorm.DB, file models.File, userBucket string) {
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

	_, err = minioClient.PutObject(ctx, userBucket, thumbPath, &thumbnailBuf, size, minio.PutObjectOptions{ContentType: "image/jpeg"})
	if err != nil {
		log.Printf("Error while uploading thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}

	thumbnail := models.Thumbnail{
		FileID:   file.ID,
		FilePath: thumbPath,
	}

	if err := db.Create(&thumbnail).Error; err != nil {
		log.Printf("Error while saving thumbnail: %s (%s) -> %s\n", thumbPath, file.FileName, err.Error())
		return
	}

	log.Printf("Thumbnail created: %s (%s)\n", thumbPath, file.FileName)
}
