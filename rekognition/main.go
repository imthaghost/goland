package main

import (
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/imthaghost/goland/rekognition/config/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rekognition"
	"github.com/corona10/goimagehash"
	"github.com/go-redis/redis/v8"
)

// Global context for Redis
var ctx = context.Background()

// Initialize AWS Rekognition client
func initRekognitionClient(awsConfig config.AWSConfig) *rekognition.Rekognition {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsConfig.Region),
		Credentials: credentials.NewStaticCredentials(awsConfig.AccessKeyID, awsConfig.SecretAccessKey, ""),
	})
	if err != nil {
		log.Fatalf("Failed to start session: %s", err)
	}
	return rekognition.New(sess)
}

// Extract keyframes using ffmpeg command and save them in outputDir
func extractKeyframes(videoPath, outputDir string) error {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.Mkdir(outputDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
	}

	cmd := exec.Command("ffmpeg", "-i", videoPath, "-vf", "select='eq(pict_type\\,I)'", "-vsync", "vfr", fmt.Sprintf("%s/frame%%03d.jpg", outputDir))
	return cmd.Run()
}

// Detect labels in a frame using AWS Rekognition
func detectLabelsWithRekognition(client *rekognition.Rekognition, filePath string) ([]*rekognition.Label, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open frame for Rekognition: %w", err)
	}
	defer file.Close()

	// Read the image into a byte slice
	imageBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read image bytes for Rekognition: %w", err)
	}

	// Call Rekognition to detect labels
	input := &rekognition.DetectLabelsInput{
		Image: &rekognition.Image{
			Bytes: imageBytes,
		},
		MaxLabels:     aws.Int64(10),
		MinConfidence: aws.Float64(80.0),
	}

	result, err := client.DetectLabels(input)
	if err != nil {
		return nil, fmt.Errorf("failed to detect labels with Rekognition: %w", err)
	}

	return result.Labels, nil
}

// Generate perceptual hash for an image at filePath
func generateHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file %s: %w", filePath, err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	hash, err := goimagehash.AverageHash(img)
	if err != nil {
		return "", fmt.Errorf("failed to generate hash: %w", err)
	}

	return hash.ToString(), nil
}

// Store hash in Redis with the associated videoID
func storeHashInRedis(client *redis.Client, hash string, videoID string) error {
	err := client.Set(ctx, hash, videoID, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store hash in Redis: %w", err)
	}
	return nil
}

// Check Redis for duplicate hashes
func checkForDuplicates(client *redis.Client, hash string) (string, error) {
	videoID, err := client.Get(ctx, hash).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // No match
	} else if err != nil {
		return "", fmt.Errorf("failed to retrieve hash from Redis: %w", err)
	}
	return videoID, nil // Match found
}

// Updated duplicate detection to use Rekognition for label analysis
func detectDuplicateFramesWithRekognition(redisClient *redis.Client, rekognitionClient *rekognition.Rekognition, frameDir, videoID string) (bool, error) {
	files, err := os.ReadDir(frameDir)
	if err != nil {
		return false, fmt.Errorf("failed to read frame directory: %w", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".jpg" {
			framePath := filepath.Join(frameDir, file.Name())

			// Step 1: Generate a perceptual hash
			hash, err := generateHash(framePath)
			if err != nil {
				return false, fmt.Errorf("error generating hash for frame %s: %w", framePath, err)
			}

			// Step 2: Check Redis for duplicate hash
			matchedVideoID, err := checkForDuplicates(redisClient, hash)
			if err != nil {
				return false, err
			}

			// Step 3: If no match in Redis, proceed with Rekognition label detection
			if matchedVideoID == "" {
				labels, err := detectLabelsWithRekognition(rekognitionClient, framePath)
				if err != nil {
					log.Printf("Failed to detect labels for %s: %v", framePath, err)
				} else {
					log.Printf("Detected labels for %s: %v", framePath, labels)
				}

				// Store the frame hash in Redis for future duplicate checks
				if err := storeHashInRedis(redisClient, hash, videoID); err != nil {
					return false, err
				}
			} else if matchedVideoID != videoID {
				return true, nil // Duplicate detected in Redis
			}
		}
	}
	return false, nil // No duplicates found
}

// Connect to Redis using config
func connectRedis(cfg config.RedisConfig) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password, // Redis password, if any
	})
	// Check if Redis connection is valid
	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %s", err)
	}
	return client
}

// Main function updated with Rekognition client and duplicate detection
func main() {
	// Load configuration
	cfg := config.New{}
	cfg.Load()
	configService := cfg.Get()

	videoPath := "./video/princess-trimmed.mp4"
	outputDir := "./video/frames"
	videoID := "t12e" // This would come from your database or video metadata

	redisClient := connectRedis(configService.RedisConfig)
	defer redisClient.Close()

	rekognitionClient := initRekognitionClient(configService.AWSConfig)

	// Step 1: Extract keyframes
	if err := extractKeyframes(videoPath, outputDir); err != nil {
		log.Fatalf("Failed to extract keyframes: %s", err)
	}

	// Step 2: Detect duplicates using Redis and Rekognition
	isDuplicate, err := detectDuplicateFramesWithRekognition(redisClient, rekognitionClient, outputDir, videoID)
	if err != nil {
		log.Fatalf("Error detecting duplicates: %s", err)
	}

	if isDuplicate {
		fmt.Println("Duplicate video detected!")
	} else {
		fmt.Println("No duplicates found.")
	}

	// Cleanup: Remove extracted frames (optional)
	if err := os.RemoveAll(outputDir); err != nil {
		log.Printf("Failed to clean up frames: %s", err)
	}
}
