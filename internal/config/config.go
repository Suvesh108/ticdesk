package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	Env           string
	DBHost        string
	DBPort        string
	DBUser        string
	DBPassword    string
	DBName        string
	DBSSLMode     string
	SessionSecret string
	SMTPHost      string
	SMTPPort      string
	SMTPFrom      string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // Load .env if present

	cfg := &Config{
		Port:          getEnv("PORT", "8080"),
		Env:           getEnv("ENV", "development"),
		DBHost:        getEnv("DB_HOST", "localhost"),
		DBPort:        getEnv("DB_PORT", "5432"),
		DBUser:        getEnv("DB_USER", "ticdesk"),
		DBPassword:    getEnv("DB_PASSWORD", "ticdesk_secret"),
		DBName:        getEnv("DB_NAME", "ticdesk"),
		DBSSLMode:     getEnv("DB_SSLMODE", "disable"),
		SessionSecret: getEnv("SESSION_SECRET", "super-secret-key-change-in-production-32bytes"),
		SMTPHost:      getEnv("SMTP_HOST", "localhost"),
		SMTPPort:      getEnv("SMTP_PORT", "1025"),
		SMTPFrom:      getEnv("SMTP_FROM", "no-reply@ticdesk.local"),
	}

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
