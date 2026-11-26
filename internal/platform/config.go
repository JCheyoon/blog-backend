package platform

import (
	"fmt"
	"os"
)

type Config struct {
	Port            string
	DatabaseURL     string
	AnthropicAPIKey string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
	}

	required := map[string]string{
		"DATABASE_URL":      cfg.DatabaseURL,
		"ANTHROPIC_API_KEY": cfg.AnthropicAPIKey,
	}
	for key, val := range required {
		if val == "" {
			return nil, fmt.Errorf("missing required environment variable: %s", key)
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
