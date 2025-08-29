package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
)

type VideoStream struct {
	Streams []struct {
		Index         int    `json:"index"`
		CodecName     string `json:"codec_name"`
		CodecLongName string `json:"codec_long_name"`
		CodecType     string `json:"codec_type"`
		Width         int    `json:"width,omitempty"`
		Height        int    `json:"height,omitempty"`
		BitRate       string `json:"bit_rate"`
		Duration      string `json:"duration"`
		Disposition   struct {
			Default      int `json:"default"`
			CleanEffects int `json:"clean_effects"`
			Captions     int `json:"captions"`
		} `json:"disposition,omitempty"`
		Tags struct {
			Language string `json:"language"`
			Title    string `json:"title"`
		} `json:"tags"`
		SampleRate   string `json:"sample_rate,omitempty"`
		Channels     int    `json:"channels,omitempty"`
		Disposition0 struct {
			Default      int `json:"default"`
			CleanEffects int `json:"clean_effects"`
		} `json:"disposition,omitempty"`
	} `json:"streams"`
}

func getVideoAspectRatio(filepath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filepath)

	var buf bytes.Buffer
	cmd.Stdout = &buf

	cmd.Run()

	var dataStruct VideoStream

	data := buf.Bytes()

	err := json.Unmarshal(data, &dataStruct)
	if err != nil {
		fmt.Printf("error unmarshalling: %v", err)
		return "", err
	}

	var Vwidth int
	var Vheight int
	for _, stream := range dataStruct.Streams {
		if stream.CodecType == "video" {
			Vwidth = stream.Width
			Vheight = stream.Height
		}
	}

	actualRatio := float64(Vwidth) / float64(Vheight)

	target16x9 := 16.0 / 9.0
	target9x16 := 9.0 / 16.0

	tolerance := 0.01

	if math.Abs(actualRatio-target16x9) < tolerance {
		return "landscape", nil
	}
	if math.Abs(actualRatio-target9x16) < tolerance {
		return "portrait", nil
	} else {
		return "other", nil
	}

}
