package main

import "os/exec"

func processVideoForFastStart(filepath string) (string, error) {
	outputPath := filepath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filepath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	cmd.Run()
	return outputPath, nil
}
