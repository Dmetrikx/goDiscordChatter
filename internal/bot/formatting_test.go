package bot

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestSendLongResponse(t *testing.T) {
	tests := []struct {
		name               string
		response           string
		wantMinChunksCount int      // Minimum chunks expected
		wantMaxChunksCount int      // Maximum chunks expected (due to AI variability)
		mockBreaks         []string // Pre-defined breaks for the mock AI
	}{
		{
			name:               "short message",
			response:           "Hello, world!",
			wantMinChunksCount: 1,
			wantMaxChunksCount: 1,
			mockBreaks:         []string{"Hello, world!"},
		},
		{
			name:               "message broken into paragraphs",
			response:           "First thought here.\n\nSecond thought here.\n\nThird thought here.",
			wantMinChunksCount: 1,
			wantMaxChunksCount: 3,
			mockBreaks:         []string{"First thought here.", "Second thought here.", "Third thought here."},
		},
		{
			name:               "exactly max length",
			response:           strings.Repeat("a", MaxDiscordMessageLength),
			wantMinChunksCount: 1,
			wantMaxChunksCount: 1,
			mockBreaks:         []string{strings.Repeat("a", MaxDiscordMessageLength)},
		},
		{
			name:               "just over max length",
			response:           strings.Repeat("a", MaxDiscordMessageLength+1),
			wantMinChunksCount: 2,
			wantMaxChunksCount: 2,
			mockBreaks:         []string{strings.Repeat("a", MaxDiscordMessageLength+1)}, // Will be split by length
		},
		{
			name:               "empty message",
			response:           "",
			wantMinChunksCount: 0,
			wantMaxChunksCount: 1,
			mockBreaks:         []string{""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock session and AI client
			mockSession := &mockDiscordSession{
				sentMessages: []string{},
			}

			mockAI := &mockAIClient{
				messageBreaks: tt.mockBreaks,
			}

			// Create a logger that discards output for tests
			logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelError, // Only log errors in tests
			}))

			bot := &Bot{
				session:  mockSession,
				aiClient: mockAI,
				logger:   logger,
			}

			ctx := context.Background()
			bot.sendLongResponse(ctx, "test-channel", tt.response)

			// Check chunk count is within expected range
			actualChunks := len(mockSession.sentMessages)
			if actualChunks < tt.wantMinChunksCount || actualChunks > tt.wantMaxChunksCount {
				t.Errorf("sendLongResponse() sent %d messages, want between %d and %d",
					actualChunks, tt.wantMinChunksCount, tt.wantMaxChunksCount)
			}

			// Verify that when concatenated, we get the original message (accounting for empty)
			// Note: for multi-chunk messages, separators (like \n\n) may be lost, so only check
			// if we expect exactly 1 chunk or if the mock breaks exactly match the response
			if tt.response != "" && tt.wantMaxChunksCount == 1 {
				concatenated := strings.Join(mockSession.sentMessages, "")
				if concatenated != tt.response {
					t.Errorf("sendLongResponse() concatenated length = %v, want %v", len(concatenated), len(tt.response))
				}
			}

			// Verify no chunk exceeds max length
			for i, msg := range mockSession.sentMessages {
				if len(msg) > MaxDiscordMessageLength {
					t.Errorf("sendLongResponse() chunk %d length = %d, exceeds max %d", i, len(msg), MaxDiscordMessageLength)
				}
			}
		})
	}
}

// mockAIClient is a mock AI client for testing
type mockAIClient struct {
	messageBreaks []string
}

func (m *mockAIClient) AskClient(ctx context.Context, prompt, systemMessage, model, provider string, maxTokens int) (string, error) {
	return "mock response", nil
}

func (m *mockAIClient) ImageOpinionOpenAI(ctx context.Context, imageURL, systemMessage, model string, maxTokens int, customPrompt *string) (string, error) {
	return "mock image opinion", nil
}

func (m *mockAIClient) ImageOpinionGrok(ctx context.Context, imageURL, systemMessage string, customPrompt *string) (string, error) {
	return "mock grok opinion", nil
}

func (m *mockAIClient) SuggestMessageBreaks(ctx context.Context, message string) ([]string, error) {
	if len(m.messageBreaks) > 0 {
		return m.messageBreaks, nil
	}
	// Default: return the message as-is
	return []string{message}, nil
}

// mockDiscordSession is a mock implementation for testing
type mockDiscordSession struct {
	sentMessages []string
}

func (m *mockDiscordSession) Open() error {
	return nil
}

func (m *mockDiscordSession) Close() error {
	return nil
}

func (m *mockDiscordSession) User(userID string, options ...discordgo.RequestOption) (*discordgo.User, error) {
	return &discordgo.User{ID: userID, Username: "testuser"}, nil
}

func (m *mockDiscordSession) ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.sentMessages = append(m.sentMessages, content)
	return &discordgo.Message{
		ID:        "msg-id",
		ChannelID: channelID,
		Content:   content,
	}, nil
}

func (m *mockDiscordSession) ChannelTyping(channelID string, options ...discordgo.RequestOption) error {
	// Mock typing - do nothing in tests
	return nil
}

func (m *mockDiscordSession) ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string, options ...discordgo.RequestOption) ([]*discordgo.Message, error) {
	return []*discordgo.Message{}, nil
}

func (m *mockDiscordSession) ChannelMessage(channelID, messageID string, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return &discordgo.Message{ID: messageID, ChannelID: channelID}, nil
}

func (m *mockDiscordSession) GuildMember(guildID, userID string, options ...discordgo.RequestOption) (*discordgo.Member, error) {
	return &discordgo.Member{User: &discordgo.User{ID: userID}}, nil
}

func (m *mockDiscordSession) AddHandler(handler interface{}) func() {
	return func() {}
}

func (m *mockDiscordSession) GetState() *discordgo.State {
	return &discordgo.State{}
}
