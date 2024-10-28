package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
)

// VideoInfo struct to parse `ffprobe` JSON output.
type VideoInfo struct {
	Streams []struct {
		Width   int    `json:"width"`
		Height  int    `json:"height"`
		Codec   string `json:"codec_name"`
		BitRate string `json:"bit_rate"`
	} `json:"streams"`
	Format struct {
		BitRate  string `json:"bit_rate"`
		Duration string `json:"duration"`
	} `json:"format"`
}

// getAdaptiveThresholds returns bitrate and codec thresholds based on resolution and video duration.
func getAdaptiveThresholds(width, height int, codec string, duration float64) (int, string) {
	log.Printf("Calculating adaptive thresholds for resolution %dx%d, codec %s, duration %.2f seconds", width, height, codec, duration)
	baseBitrate := 0

	// Determine base bitrate for given resolution and codec
	if width >= 3840 && height >= 2160 { // 4K
		baseBitrate = 30000000 // 30 Mbps for 4K
	} else if width >= 1920 && height >= 1080 { // 1080p
		baseBitrate = 8000000 // 8 Mbps for 1080p
	} else if width >= 1280 && height >= 720 { // 720p
		baseBitrate = 5000000 // 5 Mbps for 720p
	} else {
		baseBitrate = 1000000 // 1 Mbps for lower resolutions
	}

	// Adjust base bitrate down for videos longer than 5 minutes
	if duration > 300 { // Videos longer than 5 minutes
		if width >= 1920 && height >= 1080 {
			baseBitrate -= 2000000 // Lower by 2 Mbps for 1080p
		} else if width >= 1280 && height >= 720 {
			baseBitrate -= 1000000 // Lower by 1 Mbps for 720p
		} else {
			baseBitrate -= 500000 // Lower by 0.5 Mbps for lower resolutions
		}
	}

	// Cap minimum bitrate threshold to avoid excessively low values
	if baseBitrate < 1000000 {
		baseBitrate = 1000000 // Set a floor of 1 Mbps
	}

	return baseBitrate, codec
}

// checkSourceQuality validates resolution, bitrate, and codec with dynamic thresholds.
func checkSourceQuality(videoPath string) (bool, error) {
	log.Printf("Starting source quality check for video: %s", videoPath)

	// Run ffprobe command to get video metadata in JSON format.
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_format", "-show_streams", videoPath)
	var out bytes.Buffer
	cmd.Stdout = &out

	log.Println("Running ffprobe command...")
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("error running ffprobe: %v", err)
	}
	log.Println("ffprobe command executed successfully.")

	// Log the raw ffprobe output.
	log.Printf("ffprobe output: %s", out.String())

	// Parse the ffprobe JSON output.
	var videoInfo VideoInfo
	if err := json.Unmarshal(out.Bytes(), &videoInfo); err != nil {
		return false, fmt.Errorf("error parsing ffprobe output: %v", err)
	}
	log.Println("Parsed ffprobe output successfully.")

	// Extract video duration
	duration, _ := strconv.ParseFloat(videoInfo.Format.Duration, 64)

	// Check each video stream
	for i, stream := range videoInfo.Streams {
		log.Printf("Processing stream %d: resolution=%dx%d, codec=%s, bitrate=%s", i+1, stream.Width, stream.Height, stream.Codec, stream.BitRate)

		bitrate, err := strconv.Atoi(stream.BitRate)
		if err != nil {
			log.Printf("Error converting bitrate to integer: %v", err)
			continue
		}
		expectedBitrate, expectedCodec := getAdaptiveThresholds(stream.Width, stream.Height, stream.Codec, duration)

		log.Printf("Expected codec: %s, actual codec: %s", expectedCodec, stream.Codec)
		log.Printf("Expected bitrate: %d, actual bitrate: %d", expectedBitrate, bitrate)

		if stream.Codec == expectedCodec && bitrate >= expectedBitrate {
			log.Println("Stream meets source quality thresholds.")
			return true, nil
		} else {
			log.Println("Stream does not meet source quality thresholds.")
		}
	}

	log.Println("No streams met the source quality thresholds.")
	return false, nil
}

func main() {
	videoPath := "./videos/test.mp4"
	log.Println("Starting main function.")

	isSourceQuality, err := checkSourceQuality(videoPath)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}

	if isSourceQuality {
		log.Println("The video is source quality.")
		fmt.Println("The video is source quality.")
	} else {
		log.Println("The video is not source quality.")
		fmt.Println("The video is not source quality.")
	}
	log.Println("Main function completed.")
}
