package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL    string
	AuthServiceURL string
	Debug          bool
	Port           string
}

type Option func(*Config)

func WithDatabaseURL(url string) Option {
	return func(c *Config) {
		if url != "" {
			c.DatabaseURL = url
		}
	}
}

func WithAuthServiceURL(url string) Option {
	return func(c *Config) {
		if url != "" {
			c.AuthServiceURL = url
		}
	}
}

func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

func WithPort(port string) Option {
	return func(c *Config) {
		if port != "" {
			c.Port = port
		}
	}
}

func NewConfig(opts ...Option) (*Config, error) {
	envs, err := loadEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to load environment: %w", err)
	}

	c := &Config{
		DatabaseURL:    getEnvWithFallback(envs, "DATABASE_URL", ""),
		AuthServiceURL: getEnvWithFallback(envs, "AUTH_SERVICE_URL", "http://localhost:8080"),
		Debug:          getBoolEnvWithFallback(envs, "DEBUG", false),
		Port:           getEnvWithFallback(envs, "PORT", "8092"),
	}

	for _, opt := range opts {
		opt(c)
	}

	if err := c.validate(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}
	if c.AuthServiceURL == "" {
		return fmt.Errorf("AUTH_SERVICE_URL is not set")
	}
	return nil
}

func getEnvWithFallback(envs map[string]string, key, fallback string) string {
	if value, exists := envs[key]; exists && value != "" {
		return value
	}
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}

func getBoolEnvWithFallback(envs map[string]string, key string, fallback bool) bool {
	strValue := getEnvWithFallback(envs, key, strconv.FormatBool(fallback))
	boolValue, err := strconv.ParseBool(strValue)
	if err != nil {
		log.Printf("Warning: invalid boolean value for %s, using fallback", key)
		return fallback
	}
	return boolValue
}

func loadEnv() (map[string]string, error) {
	envFile := os.Getenv("ENV_FILE")
	if envFile == "" {
		_, b, _, _ := runtime.Caller(0)
		basepath := filepath.Dir(b)
		envFile = filepath.Join(basepath, "../..", ".env")
	}

	envs, err := godotenv.Read(envFile)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("Warning: .env file not found at %s, using only OS environment variables", envFile)
			return make(map[string]string), nil
		}
		return nil, fmt.Errorf("error reading .env file: %w", err)
	}
	return envs, nil
}
