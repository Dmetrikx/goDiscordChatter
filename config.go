package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration values
type Config struct {
	DiscordToken           string
	OpenAIAPIKey           string
	DiscordPoliticsChannel string
	XAIAPIKey              string
}

// LoadConfig loads environment variables from .env file and returns a Config struct
func LoadConfig() *Config {
	// Try to load .env file from the parent directory
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: .env file not found in parent directory, using environment variables")
	}

	config := &Config{
		DiscordToken:           os.Getenv("DISCORD_TOKEN"),
		OpenAIAPIKey:           os.Getenv("OPENAI_API_KEY"),
		DiscordPoliticsChannel: os.Getenv("DISCORD_POLITICS_CHANNEL"),
		XAIAPIKey:              os.Getenv("XAI_API_KEY"),
	}

	// Set default value for politics channel if not provided
	if config.DiscordPoliticsChannel == "" {
		config.DiscordPoliticsChannel = "politics"
	}

	return config
}
