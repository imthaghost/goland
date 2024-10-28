package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	// Create new thumbnail service
	thumbnailService := New()

	// Generate thumbnail
	_, err := thumbnailService.GenerateThumbnail("./videos/briannaswaybaby_1211924284_source.mp4", "./result", "briannaswaybaby_1211924284_source.mp4")
	if err != nil {
		log.Fatalf("failed to generate thumbnail: %v", err)
	}

	// Get video duration
	duration, err := thumbnailService.GetVideoDuration("./videos/briannaswaybaby_1211924284_source.mp4")
	if err != nil {
		log.Fatalf("failed to get video duration: %v", err)
	}

	log.Printf("Video duration: %.2f seconds", duration)
}

// New creates a new Thumbnail service.
func New() *Thumbnail {
	return &Thumbnail{}
}

// Thumbnail represents a service for generating GIF thumbnails from videos.
type Thumbnail struct{}

// TODO: We should do thumbnail generation and video preview generation and we will add them as different url feilds on our content struct

// GetVideoDuration retrieves the video duration in seconds using ffmpeg
func (t *Thumbnail) GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffmpeg", "-i", videoPath, "-f", "null", "-")
	log.Println("Running command: ", cmd.String())
	log.Println("Video path: ", videoPath)

	var out bytes.Buffer
	cmd.Stderr = &out // Capture ffmpeg's stderr output

	if err := cmd.Run(); err != nil {
		log.Println("Error running ffmpeg command: ", err)
		log.Println("FFmpeg output: ", out.String()) // Print the full FFmpeg output
		return 0, err
	}

	// Regex to find the Duration string in the output
	re := regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+\.\d+)`)
	matches := re.FindStringSubmatch(out.String())
	if len(matches) == 0 {
		log.Println("FFmpeg output: ", out.String()) // Log the FFmpeg output if the regex fails
		return 0, fmt.Errorf("could not extract duration from video")
	}

	// Convert the time format (HH:MM:SS) to seconds
	hours, _ := strconv.ParseFloat(matches[1], 64)
	minutes, _ := strconv.ParseFloat(matches[2], 64)
	seconds, _ := strconv.ParseFloat(matches[3], 64)

	totalSeconds := (hours * 3600) + (minutes * 60) + seconds
	return totalSeconds, nil
}

// GenerateThumbnail creates a video preview from a video file by stitching together parts from beginning, middle, and end
func (t *Thumbnail) GenerateThumbnail(videoPath, outputDir, fileName string) (string, error) {
	// Add this near the top of the GenerateThumbnail function
	log.Printf("Generating thumbnail for video: %s", videoPath)
	log.Printf("Output directory: %s", outputDir)
	log.Printf("Output file: %s", fileName)

	// Ensure the output directory exists
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Get the duration of the video
	videoDuration, err := t.GetVideoDuration(videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to get video duration: %w", err)
	}

	// Determine number of segments based on video duration
	numSegments := t.determineSegmentCount(videoDuration)

	// Calculate preview duration (10% of original, minimum 10 seconds, maximum 30 seconds)
	previewDuration := math.Min(math.Max(videoDuration*0.1, 10), 30)

	// Adjust segment duration to fit within the preview duration
	segmentDuration := previewDuration / float64(numSegments)

	log.Printf("Video duration: %.2f seconds", videoDuration)
	log.Printf("Preview duration: %.2f seconds", previewDuration)
	log.Printf("Number of segments: %d", numSegments)
	log.Printf("Segment duration: %.2f seconds", segmentDuration)

	// Calculate start times for segments
	startTimes := t.calculateStartTimes(videoDuration, numSegments, segmentDuration)

	// Generate a unique identifier
	uniqueID, err := generateUniqueID()
	if err != nil {
		return "", fmt.Errorf("failed to generate unique ID: %w", err)
	}

	// Extract the base name of the input video file (without extension)
	baseInputName := filepath.Base(videoPath)
	baseInputName = strings.TrimSuffix(baseInputName, filepath.Ext(baseInputName))

	// Temporary file paths for each segment
	tempFiles := make([]string, numSegments)
	for i := 0; i < numSegments; i++ {
		tempFiles[i] = fmt.Sprintf("%s_segment%d_%s.mp4", baseInputName, i, uniqueID)
	}

	// Determine video quality and set CRF value
	crf, err := t.determineCRF(videoPath)
	if err != nil {
		return "", fmt.Errorf("failed to determine CRF: %w", err)
	}

	// Generate individual video segments
	for i, startTime := range startTimes {
		ffmpegArgs := []string{
			"-ss", startTime,
			"-i", videoPath,
			"-t", fmt.Sprintf("%.2f", segmentDuration),
			"-vf", "scale=hd1080",
			"-c:v", "libx264",
			"-preset", "veryslow",
			"-crf", fmt.Sprintf("%d", crf),
			"-c:a", "aac",
			"-b:a", "128k",
			filepath.Join(outputDir, tempFiles[i]),
		}

		cmd := exec.Command("ffmpeg", ffmpegArgs...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to create video segment %d: %w", i+1, err)
		}
	}

	// Concatenate the video segments into a single file
	concatFileList := filepath.Join(outputDir, "concat_list.txt")

	// Write the filenames into the concat list
	var concatContent strings.Builder
	for _, tempFile := range tempFiles {
		concatContent.WriteString(fmt.Sprintf("file '%s'\n", tempFile))
	}
	if err := os.WriteFile(concatFileList, []byte(concatContent.String()), 0644); err != nil {
		return "", fmt.Errorf("failed to write concat list: %w", err)
	}

	// Log the concat list for debugging
	log.Println("Concat list content: ", concatContent.String())

	// Use absolute paths for both input and output
	absOutputPath, err := filepath.Abs(filepath.Join(outputDir, fileName))
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for output: %w", err)
	}
	absConcatFileList, err := filepath.Abs(concatFileList)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path for concat list: %w", err)
	}

	// Use ffmpeg concat demuxer to combine the segments
	ffmpegConcatArgs := []string{
		"-f", "concat",
		"-safe", "0",
		"-i", absConcatFileList,
		"-c", "copy",
		absOutputPath,
	}

	// Add this just before running the concat command
	log.Printf("Running concat command: ffmpeg %v", ffmpegConcatArgs)
	log.Printf("Working directory: %s", outputDir)

	// Change the working directory to the outputDir to ensure FFmpeg looks there for the temp files
	concatCmd := exec.Command("ffmpeg", ffmpegConcatArgs...)
	concatCmd.Dir = outputDir
	concatCmd.Stdout = os.Stdout
	concatCmd.Stderr = os.Stderr
	if err := concatCmd.Run(); err != nil {
		return "", fmt.Errorf("failed to concatenate video segments: %w", err)
	}

	// Clean up temporary files
	for _, tempFile := range tempFiles {
		os.Remove(filepath.Join(outputDir, tempFile))
	}
	os.Remove(concatFileList)

	return absOutputPath, nil
}

// determineCRF analyzes the video and returns an appropriate CRF value
func (t *Thumbnail) determineCRF(videoPath string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-count_packets",
		"-show_entries", "stream=width,height,r_frame_rate,bit_rate",
		"-of", "csv=p=0",
		videoPath)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to analyze video: %w", err)
	}

	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) < 4 {
		return 0, fmt.Errorf("unexpected ffprobe output")
	}

	width, _ := strconv.Atoi(parts[0])
	height, _ := strconv.Atoi(parts[1])
	frameRate := eval(parts[2])
	bitRate, _ := strconv.Atoi(parts[3])

	// Calculate video quality score (higher is better)
	qualityScore := float64(width*height*int(frameRate)*bitRate) / 1000000000
	log.Printf("Quality score: %.2f", qualityScore)
	// Determine CRF based on quality score
	switch {
	case qualityScore > 1000:
		log.Println("High quality, we can use higher CRF")
		return 28, nil // High quality, we can use higher CRF
	case qualityScore > 500:
		log.Println("Medium quality, we can use medium CRF")
		return 23, nil // Medium quality
	case qualityScore > 100:
		log.Println("Lower quality, we can use lower CRF")
		return 18, nil // Lower quality
	default:
		log.Println("Very low quality, we should use lower CRF to preserve quality")
		return 15, nil // Very low quality, use lower CRF to preserve quality
	}
}

// eval evaluates a fraction string like "30000/1001" to a float64
func eval(fraction string) float64 {
	parts := strings.Split(fraction, "/")
	if len(parts) != 2 {
		return 0
	}
	numerator, _ := strconv.ParseFloat(parts[0], 64)
	denominator, _ := strconv.ParseFloat(parts[1], 64)
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// generateUniqueID creates a unique identifier
func generateUniqueID() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// determineSegmentCount calculates the number of segments based on video duration
func (t *Thumbnail) determineSegmentCount(duration float64) int {
	switch {
	case duration <= 40:
		return 1
	case duration <= 120:
		return 3
	case duration <= 300:
		return 4
	default:
		return 5
	}
}

// calculateStartTimes generates start times for each segment
func (t *Thumbnail) calculateStartTimes(videoDuration float64, numSegments int, segmentDuration float64) []string {
	startTimes := make([]string, numSegments)

	if numSegments == 1 {
		startTimes[0] = "0"
	} else {
		interval := (videoDuration - segmentDuration) / float64(numSegments-1)
		for i := 0; i < numSegments; i++ {
			startTime := float64(i) * interval
			startTimes[i] = fmt.Sprintf("%.2f", startTime)
		}
	}

	return startTimes
}
