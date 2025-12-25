package bot

import (
	"context"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestSendLongResponse(t *testing.T) {
	tests := []struct {
		name            string
		response        string
		wantChunksCount int
	}{
		{
			name:            "short message",
			response:        "Hello, world!",
			wantChunksCount: 1,
		},
		{
			name:            "exactly max length",
			response:        strings.Repeat("a", MaxDiscordMessageLength),
			wantChunksCount: 1,
		},
		{
			name:            "just over max length",
			response:        strings.Repeat("a", MaxDiscordMessageLength+1),
			wantChunksCount: 2,
		},
		{
			name:            "multiple chunks",
			response:        strings.Repeat("a", MaxDiscordMessageLength*3+500),
			wantChunksCount: 4,
		},
		{
			name:            "empty message",
			response:        "",
			wantChunksCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock session
			mock := &mockDiscordSession{
				sentMessages: []string{},
			}

			bot := &Bot{
				session: mock,
			}

			ctx := context.Background()
			bot.sendLongResponse(ctx, "test-channel", tt.response)

			if len(mock.sentMessages) != tt.wantChunksCount {
				t.Errorf("sendLongResponse() sent %d messages, want %d", len(mock.sentMessages), tt.wantChunksCount)
			}

			// Verify that when concatenated, we get the original message
			concatenated := strings.Join(mock.sentMessages, "")
			if concatenated != tt.response {
				t.Errorf("sendLongResponse() concatenated = %v length, want %v length", len(concatenated), len(tt.response))
			}

			// Verify no chunk exceeds max length
			for i, msg := range mock.sentMessages {
				if len(msg) > MaxDiscordMessageLength {
					t.Errorf("sendLongResponse() chunk %d length = %d, exceeds max %d", i, len(msg), MaxDiscordMessageLength)
				}
			}
		})
	}
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
