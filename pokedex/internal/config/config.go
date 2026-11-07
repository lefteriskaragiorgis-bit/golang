package config

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv     string
	DBHost     string
	DBPort     string
	DBName     string
	DBUser     string
	DBPassword string
	DBSSLMode  string
}

func Load() (*Config, error) {
	appEnv := os.Getenv("APP_ENV")

	// Load local environment files outside production.
	if appEnv != "production" {
		for _, file := range []string{".env.local", ".env"} {
			err := godotenv.Load(file)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}

	cfg := &Config{
		AppEnv:     getEnv("APP_ENV", "local"),
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBName:     os.Getenv("DB_NAME"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBSSLMode:  os.Getenv("DB_SSLMODE"),
	}

	if cfg.DBHost == "" ||
		cfg.DBPort == "" ||
		cfg.DBName == "" ||
		cfg.DBUser == "" ||
		cfg.DBPassword == "" {
		return nil, errors.New("missing required database configuration")
	}

	if cfg.DBSSLMode == "" {
		cfg.DBSSLMode = "disable"
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}
