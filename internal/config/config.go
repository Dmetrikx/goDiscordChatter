package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	DiscordToken           string
	DiscordPoliticsChannel string
	XAIAPIKey              string
	OpenAIAPIKey           string
}

// LoadConfig loads environment variables from .env file and returns a Config struct
func LoadConfig() (*Config, error) {
	// Try to load .env file (optional - may not exist in production)
	_ = godotenv.Load(".env")

	config := &Config{
		DiscordToken:           os.Getenv("DISCORD_TOKEN"),
		DiscordPoliticsChannel: os.Getenv("DISCORD_POLITICS_CHANNEL"),
		XAIAPIKey:              os.Getenv("XAI_API_KEY"),
		OpenAIAPIKey:           os.Getenv("OPENAI_API_KEY"),
	}

	// Set default value for politics channel if not provided
	if config.DiscordPoliticsChannel == "" {
		config.DiscordPoliticsChannel = "politics"
	}

	return config, nil
}

// Validate checks that the configuration is valid
func (c *Config) Validate() error {
	if c.DiscordToken == "" {
		return NewConfigError("DISCORD_TOKEN", "environment variable is required")
	}

	if c.XAIAPIKey == "" && c.OpenAIAPIKey == "" {
		return NewConfigError("XAI_API_KEY or OPENAI_API_KEY", "at least one AI API key must be set")
	}

	if c.DiscordPoliticsChannel == "" {
		return NewConfigError("DISCORD_POLITICS_CHANNEL", "cannot be empty")
	}

	return nil
}
