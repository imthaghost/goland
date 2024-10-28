package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/imthaghost/goland/fingerprinting/cache"
	cacher "github.com/imthaghost/goland/fingerprinting/cache/redis"
	"github.com/imthaghost/goland/fingerprinting/config"

	"github.com/corona10/goimagehash"
	"gocv.io/x/gocv"
)

// Constants
const (
	frameIntervalSec        = 1   // Seconds between frames
	hashSimilarityThreshold = 0.6 // 60% similarity for duplicates
	cacheExpiration         = 24 * time.Hour * 7
)

// FingerPrint represents a service for generating perceptual hashes from video frames
type FingerPrint struct {
	Cache cache.Service
}

func New(cache cache.Service) *FingerPrint {
	return &FingerPrint{
		Cache: cache,
	}
}

func (f *FingerPrint) GenerateVideoFingerprint(videoPath string) ([]string, error) {
	video, err := gocv.VideoCaptureFile(videoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open video file: %v", err)
	}
	defer video.Close()

	var hashes []string
	frameInterval := int(video.Get(gocv.VideoCaptureFPS))

	for frameCount := 0; frameCount < int(video.Get(gocv.VideoCaptureFrameCount)); frameCount += frameInterval {
		frame := gocv.NewMat()
		if !video.Read(&frame) || frame.Empty() {
			continue
		}

		hash, err := f.GenerateImageHash(frame)
		if err == nil {
			hashes = append(hashes, hash)
		}
		frame.Close()
	}

	fingerprint := cache.VideoFingerprint{
		Hashes:    hashes,
		Timestamp: time.Now(),
	}

	key := "video:" + generateVideoKey(videoPath)
	_, err = f.Cache.Set(context.Background(), key, fingerprint, time.Hour*24*7)
	if err != nil {
		return nil, fmt.Errorf("failed to store fingerprint: %w", err)
	}

	return hashes, nil
}

// CheckForDuplicate checks for duplicate video based on hash similarity
func (f *FingerPrint) CheckForDuplicate(videoPath string) (bool, error) {
	key := "video:" + generateVideoKey(videoPath)

	// First, attempt to retrieve an existing fingerprint from the cache
	existingFingerprint, err := f.Cache.Get(context.Background(), key)
	if err != nil {
		return false, fmt.Errorf("failed to retrieve fingerprint: %w", err)
	}

	// If a fingerprint already exists in the cache, calculate similarity
	if existingFingerprint != nil {
		// Generate a new fingerprint for comparison
		currentHashes, err := f.GenerateVideoFingerprint(videoPath)
		if err != nil {
			return false, fmt.Errorf("failed to generate fingerprint: %w", err)
		}

		similarity := calculateFingerprintSimilarity(existingFingerprint.Hashes, currentHashes)
		return similarity >= hashSimilarityThreshold, nil
	}

	// If no existing fingerprint, generate and store a new one
	_, err = f.GenerateVideoFingerprint(videoPath)
	if err != nil {
		return false, fmt.Errorf("failed to store new fingerprint: %w", err)
	}

	return false, nil // New video, not a duplicate
}

// GenerateImageHash generates a perceptual hash from an image frame
func (f *FingerPrint) GenerateImageHash(img gocv.Mat) (string, error) {
	if img.Empty() {
		return "", fmt.Errorf("empty frame, unable to generate hash")
	}

	// Convert Mat to image.Image directly without PNG encoding/decoding
	rows := img.Rows()
	cols := img.Cols()

	// Create a new RGBA image
	bounds := image.Rect(0, 0, cols, rows)
	rgbaImg := image.NewRGBA(bounds)

	// Copy the pixel data from Mat to RGBA image
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			pixel := img.GetVecbAt(y, x)
			rgbaImg.Set(x, y, color.RGBA{
				B: pixel[0], // OpenCV uses BGR format
				G: pixel[1],
				R: pixel[2],
				A: 255,
			})
		}
	}

	// Generate perceptual hash using goimagehash
	hash, err := goimagehash.PerceptionHash(rgbaImg)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash: %v", err)
	}

	hashStr := hash.ToString()
	log.Printf("Generated hash: %s", hashStr)
	return hashStr, nil
}

// calculateFingerprintSimilarity calculates similarity between two sets of hashes
func calculateFingerprintSimilarity(hashes1, hashes2 []string) float64 {
	if len(hashes1) == 0 || len(hashes2) == 0 {
		return 0.0
	}

	matches := 0
	for i := range hashes1 {
		if i >= len(hashes2) {
			break
		}

		// Remove the 'p:' prefix if it exists
		h1Str := strings.TrimPrefix(hashes1[i], "p:")
		h2Str := strings.TrimPrefix(hashes2[i], "p:")

		// Convert hex string to uint64
		h1Val, err := strconv.ParseUint(h1Str, 16, 64)
		if err != nil {
			log.Printf("Error parsing hash1: %v", err)
			continue
		}

		h2Val, err := strconv.ParseUint(h2Str, 16, 64)
		if err != nil {
			log.Printf("Error parsing hash2: %v", err)
			continue
		}

		// Create ImageHash objects
		h1 := goimagehash.NewImageHash(h1Val, goimagehash.PHash)
		h2 := goimagehash.NewImageHash(h2Val, goimagehash.PHash)

		distance, err := h1.Distance(h2)
		if err != nil {
			log.Printf("Error calculating distance: %v", err)
			continue
		}

		if distance <= 10 { // Threshold for frame similarity
			matches++
		}
	}

	if matches == 0 {
		return 0.0
	}

	return float64(matches) / float64(len(hashes1))
}

func generateVideoKey(videoPath string) string {
	hasher := sha256.New()
	hasher.Write([]byte(videoPath))
	videoHash := hex.EncodeToString(hasher.Sum(nil))
	return videoHash[:12]
}

func main() {
	// load configService
	cfg := config.New{}
	cfg.Load()
	configService := cfg.Get()

	// Initialize Redis cache
	cacheService := cacher.New(configService)

	// Initialize fingerprint service
	fingerprintService := New(cacheService)
	videoPath := "./video/trimmed-dupe.mp4"

	log.Printf("Starting fingerprint process for video: %s", videoPath)

	// Check for duplicates using the service's built-in method
	isDuplicate, err := fingerprintService.CheckForDuplicate(videoPath)
	if err != nil {
		log.Printf("Error checking for duplicates: %v", err)
	}

	if isDuplicate {
		log.Println("Duplicate video detected.")
	} else {
		log.Println("New video processed and fingerprints saved.")
	}
}
