package config

import (
	"log"
	"os"
)

type Config struct {
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	FromEmail    string
	AppEnv       string
}

func LoadConfig() *Config {
	cfg := &Config{
		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnv("SMTP_PORT", ""),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		FromEmail:    getEnv("FROM_EMAIL", ""),
		AppEnv:       getEnv("APP_ENV", ""),
	}

	if cfg.SMTPUser == "" || cfg.SMTPPassword == "" {
		log.Fatal("SMTP_USER e SMTP_PASSWORD são obrigatórios")
	}

	return cfg
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
