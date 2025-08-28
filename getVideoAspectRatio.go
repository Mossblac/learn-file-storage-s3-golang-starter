package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func getVideoAspectRatio(filepath string) (string, error) {
	arguments := fmt.Sprintf("-v error -print_format json -show_streams %v", filepath)
	cmd := exec.Command("ffprobe", arguments)

	var buf bytes.Buffer
	cmd.Stdout = &buf

	cmd.Run()

	type Streams struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	}

	var streams Streams

	data := buf.Bytes()

	err := json.Unmarshal(data, &streams)
	if err != nil {
		return "", err
	}

	GCD := gcd(streams.Width, streams.Height)
	aspectFirst := streams.Width * GCD
	aspectSecond := streams.Height * GCD

	return fmt.Sprintf("%v:%v", aspectFirst, aspectSecond), nil

}
