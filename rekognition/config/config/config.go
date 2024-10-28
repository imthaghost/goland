package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

const (
	DEV     = "dev"
	STAGING = "staging"
	PROD    = "prod"
)

// New will create a new config
type New struct{}

// Load will load the config
func (n *New) Load() {
	err := godotenv.Load()
	if err != nil {
		log.Println(err)
	}
}

// Get will return the default config
func (n *New) Get() Config {
	return Config{
		General:     getGeneralConfig(),
		RedisConfig: getRedisConfig(),
		AWSConfig:   getAWSConfig(),
	}
}

// getGeneralConfig returns the general config
func getGeneralConfig() GeneralConfig {
	// default
	config := GeneralConfig{
		AppEnv:       os.Getenv("APP_ENV"),
		DisplayToken: os.Getenv("DISPLAY_TOKEN"),
	}

	if config.AppEnv == "" {
		config.AppEnv = DEV
		config.BaseURL = "http://localhost:8080"
	}

	if config.AppEnv == PROD {
		config.BaseURL = "https://api.contrihub.lol"
	}

	return config
}

func getRedisConfig() RedisConfig {
	// default
	config := RedisConfig{
		Host:     os.Getenv("REDIS_HOST"),
		Port:     os.Getenv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
	}

	return config
}

// AWSConfig contains information that allows us to interact with the AWS API
func getAWSConfig() AWSConfig {
	// default
	config := AWSConfig{
		AccessKeyID:     os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:          os.Getenv("AWS_REGION"),
	}

	return config
}
