package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func getVideoAspectRatio(filePath string) (string, error) {
	command := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var b bytes.Buffer
	command.Stdout = &b
	err := command.Run()
	if err != nil {
		return "", err
	}
	type RatioStruct struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	newRatio := RatioStruct{}
	if err := json.Unmarshal(b.Bytes(), &newRatio); err != nil {
		return "", err
	}
	if len(newRatio.Streams) == 0 {
		return "", fmt.Errorf("no videos stream found")
	}

	ratio := float64(newRatio.Streams[0].Width) / float64(newRatio.Streams[0].Height)

	if ratio > 1.7 && ratio < 1.8 {
		return "16:9", nil
	}
	if ratio > 0.5 && ratio < 0.6 {
		return "9:16", nil
	}
	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := fmt.Sprintf("%s.processing", filePath)
	command := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	err := command.Run()
	if err != nil {
		return "", err
	}
	return outputPath, err
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignedClient := s3.NewPresignClient(s3Client)
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	presignedObject, err := presignedClient.PresignGetObject(context.Background(), input, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}
	return presignedObject.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}
	inputs := strings.Split(*video.VideoURL, ",")
	presignedURL, err := generatePresignedURL(cfg.s3Client, inputs[0], inputs[1], 15*time.Minute)
	if err != nil {
		return database.Video{}, err
	}
	video.VideoURL = &presignedURL
	return video, nil
}
