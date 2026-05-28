package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
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
