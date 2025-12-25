package config

import (
	"os"
	"testing"
)

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with both API keys",
			config: &Config{
				DiscordToken:           "test-discord-token",
				DiscordPoliticsChannel: "politics",
				XAIAPIKey:              "test-xai-key",
				OpenAIAPIKey:           "test-openai-key",
			},
			wantErr: false,
		},
		{
			name: "valid config with only XAI key",
			config: &Config{
				DiscordToken:           "test-discord-token",
				DiscordPoliticsChannel: "politics",
				XAIAPIKey:              "test-xai-key",
			},
			wantErr: false,
		},
		{
			name: "valid config with only OpenAI key",
			config: &Config{
				DiscordToken:           "test-discord-token",
				DiscordPoliticsChannel: "politics",
				OpenAIAPIKey:           "test-openai-key",
			},
			wantErr: false,
		},
		{
			name: "missing Discord token",
			config: &Config{
				DiscordPoliticsChannel: "politics",
				XAIAPIKey:              "test-xai-key",
			},
			wantErr: true,
			errMsg:  "DISCORD_TOKEN",
		},
		{
			name: "missing all AI keys",
			config: &Config{
				DiscordToken:           "test-discord-token",
				DiscordPoliticsChannel: "politics",
			},
			wantErr: true,
			errMsg:  "XAI_API_KEY or OPENAI_API_KEY",
		},
		{
			name: "empty politics channel",
			config: &Config{
				DiscordToken:           "test-discord-token",
				DiscordPoliticsChannel: "",
				XAIAPIKey:              "test-xai-key",
			},
			wantErr: true,
			errMsg:  "DISCORD_POLITICS_CHANNEL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Save original env vars
	origDiscordToken := os.Getenv("DISCORD_TOKEN")
	origXAIKey := os.Getenv("XAI_API_KEY")
	origOpenAIKey := os.Getenv("OPENAI_API_KEY")
	origPoliticsChannel := os.Getenv("DISCORD_POLITICS_CHANNEL")

	// Cleanup after test
	defer func() {
		os.Setenv("DISCORD_TOKEN", origDiscordToken)
		os.Setenv("XAI_API_KEY", origXAIKey)
		os.Setenv("OPENAI_API_KEY", origOpenAIKey)
		os.Setenv("DISCORD_POLITICS_CHANNEL", origPoliticsChannel)
	}()

	tests := []struct {
		name          string
		envVars       map[string]string
		wantToken     string
		wantXAIKey    string
		wantOpenAIKey string
		wantPolitics  string
	}{
		{
			name: "loads all env vars",
			envVars: map[string]string{
				"DISCORD_TOKEN":            "test-token",
				"XAI_API_KEY":              "test-xai",
				"OPENAI_API_KEY":           "test-openai",
				"DISCORD_POLITICS_CHANNEL": "test-politics",
			},
			wantToken:     "test-token",
			wantXAIKey:    "test-xai",
			wantOpenAIKey: "test-openai",
			wantPolitics:  "test-politics",
		},
		{
			name: "default politics channel",
			envVars: map[string]string{
				"DISCORD_TOKEN": "test-token",
				"XAI_API_KEY":   "test-xai",
			},
			wantToken:    "test-token",
			wantXAIKey:   "test-xai",
			wantPolitics: "politics",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear env vars
			os.Unsetenv("DISCORD_TOKEN")
			os.Unsetenv("XAI_API_KEY")
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("DISCORD_POLITICS_CHANNEL")

			// Set test env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := LoadConfig()
			if err != nil {
				t.Fatalf("LoadConfig() error = %v", err)
			}

			if cfg.DiscordToken != tt.wantToken {
				t.Errorf("DiscordToken = %v, want %v", cfg.DiscordToken, tt.wantToken)
			}
			if cfg.XAIAPIKey != tt.wantXAIKey {
				t.Errorf("XAIAPIKey = %v, want %v", cfg.XAIAPIKey, tt.wantXAIKey)
			}
			if cfg.OpenAIAPIKey != tt.wantOpenAIKey {
				t.Errorf("OpenAIAPIKey = %v, want %v", cfg.OpenAIAPIKey, tt.wantOpenAIKey)
			}
			if cfg.DiscordPoliticsChannel != tt.wantPolitics {
				t.Errorf("DiscordPoliticsChannel = %v, want %v", cfg.DiscordPoliticsChannel, tt.wantPolitics)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
