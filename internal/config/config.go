// internal/config/config.go
package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	// Cloudflare R2
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2PublicURL       string // e.g. https://pub-xxx.r2.dev
}

func Load() *Config {
	// .env is optional — Railway injects env vars directly
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, using environment variables")
	}

	return &Config{
		Port:              getEnv("PORT", "8080"),
		DatabaseURL:       mustEnv("DATABASE_URL"),
		JWTSecret:         mustEnv("JWT_SECRET"),
		R2AccountID:       mustEnv("R2_ACCOUNT_ID"),
		R2AccessKeyID:     mustEnv("R2_ACCESS_KEY_ID"),
		R2SecretAccessKey: mustEnv("R2_SECRET_ACCESS_KEY"),
		R2BucketName:      mustEnv("R2_BUCKET_NAME"),
		R2PublicURL:       mustEnv("R2_PUBLIC_URL"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %q is not set", key)
	}
	return v
}
