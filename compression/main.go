package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func main() {
	inputFile := "./videos/started_compressed.mp4"
	outputFile := "started_compressed.mp4"
	outputDir := "./result"

	originalSize, err := getFileSize(inputFile)
	if err != nil {
		log.Fatalf("failed to get original file size: %v", err)
	}
	log.Printf("Original file size: %.2f MB", float64(originalSize)/1e6)

	videoQuality, err := getVideoQuality(inputFile)
	if err != nil {
		log.Fatalf("failed to get video quality: %v", err)
	}

	// Determine dynamic CRF based on video quality
	crf := determineCRF(videoQuality, originalSize)
	log.Printf("Using CRF value: %d", crf)

	// Record the start time
	startTime := time.Now()

	err = compressVideo(inputFile, outputDir, outputFile, crf)
	if err != nil {
		log.Fatalf("failed to compress video: %v", err)
	}

	// Calculate the elapsed time
	elapsedTime := time.Since(startTime)

	compressedSize, err := getFileSize(filepath.Join(outputDir, outputFile))
	if err != nil {
		log.Fatalf("failed to get compressed file size: %v", err)
	}
	log.Printf("Compressed file size: %.2f MB", float64(compressedSize)/1e6)

	// Calculate and log the compression percentage
	compressionPercentage := (float64(originalSize-compressedSize) / float64(originalSize)) * 100
	log.Printf("Compression percentage: %.2f%%", compressionPercentage)

	// Log the time taken for compression
	log.Printf("Time taken for compression: %v", elapsedTime)
}

func getVideoQuality(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=bit_rate", "-of", "default=noprint_wrappers=1:nokey=1", filePath)

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to run ffprobe: %v", err)
	}

	bitrate, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return "", err
	}

	// Determine video quality based on bitrate
	if bitrate > 2000000 { // Example threshold for high quality
		return "high", nil
	} else if bitrate > 1000000 { // Example threshold for medium quality
		return "medium", nil
	} else {
		return "low", nil
	}
}

// determineCRF determines the Constant Rate Factor (CRF) value based on video quality and file size
func determineCRF(videoQuality string, fileSize int64) int {
	switch videoQuality {
	case "high":
		return 23 // Less aggressive compression
	case "medium":
		return 28 // Moderate compression
	case "low":
		// Consider files 890 MB or larger as large files
		if fileSize >= 890*1024*1024 { // 890 MB in bytes
			return 30 // Higher CRF for large low-quality files
		}
		return 28
	default:
		return 23 // Fallback value
	}
}

func compressVideo(inputFile, outputDir, outputFileName string, crf int) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	outputFilePath := filepath.Join(outputDir, outputFileName)
	cmd := exec.Command("ffmpeg", "-i", inputFile, "-vcodec", "libx264", "-crf", strconv.Itoa(crf), "-preset", "medium", "-acodec", "aac", "-b:a", "128k", "-movflags", "+faststart", outputFilePath)

	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	return fmt.Errorf("failed to get stdout pipe: %v", err)
	//}
	//stderr, err := cmd.StderrPipe()
	//if err != nil {
	//	return fmt.Errorf("failed to get stderr pipe: %v", err)
	//}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	//go logOutput(stdout, "FFmpeg stdout")
	//go logOutput(stderr, "FFmpeg stderr")

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("FFmpeg failed: %v", err)
	}

	return nil
}

func logOutput(reader io.Reader, prefix string) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		log.Printf("[%s] %s", prefix, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Printf("[%s] error reading output: %v", prefix, err)
	}
}

func getFileSize(filePath string) (int64, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0, err
	}
	return fileInfo.Size(), nil
}
