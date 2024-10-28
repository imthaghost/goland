package config

// Service is an interface that defines the functions needed to implement a Config Service.
type Service interface {
	// Load will do any config setup (like load env vars)
	Load()
	// Get will get the config
	Get() Config
}

// Config is a service that is designed to provide various configuration to the rest of the application.
type Config struct {
	General     GeneralConfig
	AWSConfig   AWSConfig
	RedisConfig RedisConfig
}

// GeneralConfig contains general information that the service needs to run.
type GeneralConfig struct {
	AppEnv       string // the environment that the application is running in (dev, prod, etc)
	DisplayToken string // the token used to display trending channels on the landing page
	BaseURL      string // the base URL of the application
}

// RedisConfig contains information that allows us to interact with the Redis API
type RedisConfig struct {
	Host     string
	Port     string
	Password string
}

// AWSConfig contains information that allows us to interact with the AWS API
type AWSConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
}
