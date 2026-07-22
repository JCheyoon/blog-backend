package platform

import (
	"encoding/base64"
	"fmt"
	"os"
)

type Config struct {
	Port              string
	DatabaseURL       string
	JWTSecret         string
	AdminEmail        string
	AdminPasswordHash string 
	AnthropicAPIKey   string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AdminEmail:      os.Getenv("ADMIN_EMAIL"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}

	passwordHashB64 := os.Getenv("ADMIN_PASSWORD_HASH_B64")

	required := map[string]string{
		"DATABASE_URL":            cfg.DatabaseURL,
		"JWT_SECRET":              cfg.JWTSecret,
		"ADMIN_EMAIL":             cfg.AdminEmail,
		"ADMIN_PASSWORD_HASH_B64": passwordHashB64,
		"ANTHROPIC_API_KEY":       cfg.AnthropicAPIKey,
	}
	for key, val := range required {
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", key)
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(passwordHashB64)
	if err != nil {
		return nil, fmt.Errorf("ADMIN_PASSWORD_HASH_B64 is not valid base64: %w", err)
	}
	cfg.AdminPasswordHash = string(decoded)

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}