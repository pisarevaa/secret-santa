package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	BaseURL      string
	DatabasePath string
	ResendAPIKey string
	EmailFrom    string
	Port         int
	Env          string // "production" | "development"
	LogLevel     string
}

func Load() (*Config, error) {
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		port = p
	}

	env := os.Getenv("ENV")
	if env == "" {
		env = "development"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "app.db"
	}

	return &Config{
		BaseURL:      os.Getenv("BASE_URL"),
		DatabasePath: dbPath,
		ResendAPIKey: os.Getenv("RESEND_API_KEY"),
		EmailFrom:    os.Getenv("EMAIL_FROM"),
		Port:         port,
		Env:          env,
		LogLevel:     logLevel,
	}, nil
}

func (c *Config) IsDev() bool {
	return c.Env == "development"
}
