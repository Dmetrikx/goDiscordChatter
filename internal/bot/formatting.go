package bot

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// sendLongResponse sends responses broken up into natural chunks with human-like timing
// Uses AI to determine natural breaking points and adds realistic delays between messages
func (b *Bot) sendLongResponse(ctx context.Context, channelID, response string) {
	// Get AI-suggested message breaks
	chunks, err := b.aiClient.SuggestMessageBreaks(ctx, response)
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to get message breaks, using fallback",
			"error", err)
		// Fallback to simple chunking
		chunks = b.fallbackChunking(response)
	}

	// Send each chunk with human-like delays
	for i, chunk := range chunks {
		// Ensure chunk doesn't exceed Discord limit
		if len(chunk) > MaxDiscordMessageLength {
			// If a chunk is too long, split it further
			subChunks := b.splitLongChunk(chunk)
			for j, subChunk := range subChunks {
				b.sendChunkWithDelay(ctx, channelID, subChunk, i > 0 || j > 0)
			}
		} else {
			b.sendChunkWithDelay(ctx, channelID, chunk, i > 0)
		}
	}
}

// sendChunkWithDelay sends a single chunk with optional typing delay before it
func (b *Bot) sendChunkWithDelay(ctx context.Context, channelID, chunk string, addDelay bool) {
	if addDelay {
		// Calculate a human-like delay based on chunk length
		delay := b.calculateTypingDelay(chunk)

		b.logger.InfoContext(ctx, "waiting before next message chunk",
			"delay_ms", delay.Milliseconds(),
			"chunk_length", len(chunk))

		// Show typing indicator while "typing"
		// Discord typing indicator lasts ~10 seconds, so we trigger it periodically
		go b.showTypingIndicator(ctx, channelID, delay)

		time.Sleep(delay)
	}

	_, err := b.session.ChannelMessageSend(channelID, chunk)
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send message chunk",
			"channel_id", channelID,
			"error", err)
	}
}

// showTypingIndicator displays the typing indicator for the duration of the delay
func (b *Bot) showTypingIndicator(ctx context.Context, channelID string, duration time.Duration) {
	// Discord's typing indicator lasts ~10 seconds, so we need to refresh it for longer delays
	ticker := time.NewTicker(8 * time.Second) // Refresh every 8 seconds to be safe
	defer ticker.Stop()

	// Send initial typing indicator
	err := b.session.ChannelTyping(channelID)
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send typing indicator", "error", err)
		return
	}

	// Keep typing indicator alive for the full duration
	timeout := time.After(duration)
	for {
		select {
		case <-timeout:
			return
		case <-ticker.C:
			err := b.session.ChannelTyping(channelID)
			if err != nil {
				b.logger.ErrorContext(ctx, "failed to refresh typing indicator", "error", err)
				return
			}
		}
	}
}

// calculateTypingDelay calculates a realistic delay based on message length
// Simulates the time it would take a human to type and send the message
func (b *Bot) calculateTypingDelay(chunk string) time.Duration {
	// Base delay on chunk length (simulate typing speed)
	chunkLength := len(chunk)
	typingTime := time.Duration(chunkLength) * TypingSpeed / 80 // Adjusted for more realistic feel

	// Add some variance and clamp to reasonable bounds
	delay := MinMessageDelay + typingTime
	if delay > MaxMessageDelay {
		delay = MaxMessageDelay
	}

	return delay
}

// fallbackChunking provides simple paragraph-based chunking when AI breaks fail
func (b *Bot) fallbackChunking(response string) []string {
	chunks := []string{}
	paragraphs := strings.Split(response, "\n\n")

	currentChunk := ""
	for _, para := range paragraphs {
		para = strings.TrimSpace(para)
		if para == "" {
			continue
		}

		// If adding this paragraph would make chunk too large, start new chunk
		if len(currentChunk) > 0 && len(currentChunk)+len(para)+2 > 800 {
			chunks = append(chunks, strings.TrimSpace(currentChunk))
			currentChunk = para
		} else {
			if len(currentChunk) > 0 {
				currentChunk += "\n\n" + para
			} else {
				currentChunk = para
			}
		}
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, strings.TrimSpace(currentChunk))
	}

	// If we ended up with just one or no chunks, return original
	if len(chunks) <= 1 {
		return []string{response}
	}

	return chunks
}

// splitLongChunk splits a chunk that exceeds Discord's limit
func (b *Bot) splitLongChunk(chunk string) []string {
	subChunks := []string{}

	for i := 0; i < len(chunk); i += MaxDiscordMessageLength {
		end := int(math.Min(float64(i+MaxDiscordMessageLength), float64(len(chunk))))
		subChunks = append(subChunks, chunk[i:end])
	}

	return subChunks
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
