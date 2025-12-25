package bot

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// sendLongResponse sends long responses in chunks to respect Discord's message length limit
func (b *Bot) sendLongResponse(ctx context.Context, channelID, response string) {
	for i := 0; i < len(response); i += MaxDiscordMessageLength {
		end := i + MaxDiscordMessageLength
		if end > len(response) {
			end = len(response)
		}

		chunk := response[i:end]
		_, err := b.session.ChannelMessageSend(channelID, chunk)
		if err != nil {
			b.logger.ErrorContext(ctx, "failed to send message chunk",
				"channel_id", channelID,
				"chunk_index", i/MaxDiscordMessageLength,
				"error", err)
		}
	}
}

// formatChannelHistory fetches and formats recent messages
func (b *Bot) formatChannelHistory(ctx context.Context, channelID string, numMessages int) (string, error) {
	messages, err := b.session.ChannelMessages(channelID, numMessages, "", "", "")
	if err != nil {
		return "", fmt.Errorf("failed to fetch channel messages: %w", err)
	}

	// Reverse the messages to show oldest first
	var formatted []string
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		displayName := getDisplayName(b.session, msg)
		formatted = append(formatted, fmt.Sprintf("%s: %s", displayName, msg.Content))
	}

	return strings.Join(formatted, "\n"), nil
}

// getDisplayName retrieves the display name for a message author
func getDisplayName(session interface {
	GuildMember(guildID, userID string, options ...discordgo.RequestOption) (*discordgo.Member, error)
}, msg *discordgo.Message) string {
	displayName := msg.Author.Username
	if msg.GuildID != "" {
		member, err := session.GuildMember(msg.GuildID, msg.Author.ID)
		if err == nil && member.Nick != "" {
			displayName = member.Nick
		}
	}
	return displayName
}

// extractProviderAndArgs extracts provider from arguments and returns remaining args
func extractProviderAndArgs(args []string, defaultProvider string) (string, []string) {
	provider := defaultProvider
	if len(args) > 0 {
		lower := strings.ToLower(args[0])
		// Check if first arg is a known provider
		if lower == "grok" || lower == "openai" {
			provider = lower
			args = args[1:]
		}
	}
	return provider, args
}

// providerDisplayName returns a formatted display name for a provider
func providerDisplayName(provider string) string {
	switch provider {
	case "grok":
		return "Grok"
	case "openai":
		return "OpenAI"
	default:
		return strings.Title(provider)
	}
}

// modelVersion returns the version string for a given provider and model
func modelVersion(provider, model string) string {
	// Import from ai package would be needed here
	// For now, we'll handle this in the bot file
	return ""
}
