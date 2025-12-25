package discord

import (
	"github.com/bwmarrin/discordgo"
)

// Session defines the interface for Discord session operations
type Session interface {
	// Open opens a websocket connection to Discord
	Open() error

	// Close closes the websocket connection to Discord
	Close() error

	// User returns the current user
	User(userID string, options ...discordgo.RequestOption) (*discordgo.User, error)

	// ChannelMessageSend sends a message to a channel
	ChannelMessageSend(channelID string, content string, options ...discordgo.RequestOption) (*discordgo.Message, error)

	// ChannelMessages retrieves messages from a channel
	ChannelMessages(channelID string, limit int, beforeID, afterID, aroundID string, options ...discordgo.RequestOption) ([]*discordgo.Message, error)

	// ChannelMessage retrieves a specific message from a channel
	ChannelMessage(channelID, messageID string, options ...discordgo.RequestOption) (*discordgo.Message, error)

	// GuildMember retrieves a guild member
	GuildMember(guildID, userID string, options ...discordgo.RequestOption) (*discordgo.Member, error)

	// AddHandler adds an event handler
	AddHandler(handler interface{}) func()

	// GetState returns the session state
	GetState() *discordgo.State
}

// DiscordSession wraps discordgo.Session to implement the Session interface
type DiscordSession struct {
	*discordgo.Session
}

// NewDiscordSession creates a new DiscordSession wrapper
func NewDiscordSession(token string) (*DiscordSession, error) {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, err
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	return &DiscordSession{Session: session}, nil
}

// GetState returns the session state
func (d *DiscordSession) GetState() *discordgo.State {
	return d.State
}
