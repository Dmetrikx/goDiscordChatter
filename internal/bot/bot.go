package bot

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/Dmetrikx/goDiscordChatter/internal/ai"
	"github.com/Dmetrikx/goDiscordChatter/internal/config"
	"github.com/Dmetrikx/goDiscordChatter/internal/discord"
)

// Bot represents the Discord bot
type Bot struct {
	session  discord.Session
	aiClient ai.Client
	config   *config.Config
	logger   *slog.Logger
}

// NewBot creates a new bot instance
func NewBot(cfg *config.Config, logger *slog.Logger) (*Bot, error) {
	session, err := discord.NewDiscordSession(cfg.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	aiClient := ai.NewAIClient(cfg.OpenAIAPIKey, cfg.XAIAPIKey, logger)

	bot := &Bot{
		session:  session,
		aiClient: aiClient,
		config:   cfg,
		logger:   logger,
	}

	// Register message handler
	session.AddHandler(bot.messageHandler)

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start(ctx context.Context) error {
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	user, err := b.session.User("@me")
	if err != nil {
		return fmt.Errorf("error obtaining account details: %w", err)
	}

	b.logger.InfoContext(ctx, "bot started",
		"username", user.Username,
		"user_id", user.ID)

	return nil
}

// Close closes the bot session
func (b *Bot) Close(ctx context.Context) error {
	b.logger.InfoContext(ctx, "closing bot session")
	return b.session.Close()
}

// messageHandler handles incoming messages
func (b *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	ctx := context.Background()

	// Ignore messages from the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Check if message starts with command prefix
	if !strings.HasPrefix(m.Content, "!") {
		return
	}

	// Parse command and arguments
	parts := strings.Fields(m.Content)
	if len(parts) == 0 {
		return
	}

	command := strings.TrimPrefix(parts[0], "!")
	args := parts[1:]

	b.logger.InfoContext(ctx, "received command",
		"command", command,
		"user_id", m.Author.ID,
		"username", m.Author.Username,
		"channel_id", m.ChannelID,
		"args_count", len(args))

	// Route to appropriate command handler
	switch command {
	case "ping":
		b.handlePing(ctx, s, m)
	case "ask":
		b.handleAsk(ctx, s, m, args)
	case "opinion":
		b.handleOpinion(ctx, s, m, args)
	case "who_won":
		b.handleWhoWon(ctx, s, m, args)
	case "user_opinion":
		b.handleUserOpinion(ctx, s, m, args)
	case "most":
		b.handleMost(ctx, s, m, args)
	case "image_opinion":
		b.handleImageOpinion(ctx, s, m, args)
	case "roast":
		b.handleRoast(ctx, s, m, args)
	default:
		b.logger.InfoContext(ctx, "unknown command", "command", command)
	}
}

// handlePing responds with "Pong!"
func (b *Bot) handlePing(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate) {
	_, err := s.ChannelMessageSend(m.ChannelID, "Pong!")
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to send ping response", "error", err)
	}
}

// handleAsk handles the !ask command
func (b *Bot) handleAsk(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !ask [grok|openai] <question>")
		return
	}

	provider, args := extractProviderAndArgs(args, ai.DefaultProvider)
	prompt := strings.Join(args, " ")

	model := ai.DefaultGrokModel
	persona := ai.GrokPersona
	if provider == ai.ProviderOpenAI {
		model = ai.DefaultOpenAIModel
		persona = ai.OpenAIPersona
	}

	b.sendThinkingMessage(ctx, s, m.ChannelID, provider, model)

	response, err := b.aiClient.AskClient(ctx, prompt, persona, model, provider, ai.DefaultMaxTokens)
	if err != nil {
		b.logger.ErrorContext(ctx, "AI request failed",
			"command", "ask",
			"provider", provider,
			"error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(ctx, m.ChannelID, response)
}

// handleOpinion handles the !opinion command
func (b *Bot) handleOpinion(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "Let me think about what everyone has been saying...")

	provider, args := extractProviderAndArgs(args, ai.DefaultProvider)
	model := ai.DefaultGrokModel
	persona := ai.GrokPersona
	if provider == ai.ProviderOpenAI {
		model = ai.DefaultOpenAIModel
		persona = ai.OpenAIPersona
	}

	numMessages := DefaultHistoryMessageCount
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			numMessages = n
		}
	}

	contextStr, err := b.formatChannelHistory(ctx, m.ChannelID, numMessages)
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to fetch channel history", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	systemMessage := fmt.Sprintf("%s\nHere are the last %d messages in this channel:\n%s\n"+
		"Form an opinion or summary about the conversation.", persona, numMessages, contextStr)

	response, err := b.aiClient.AskClient(ctx, "What is your opinion on the recent conversation?",
		systemMessage, model, provider, ai.DefaultMaxTokens)
	if err != nil {
		b.logger.ErrorContext(ctx, "AI request failed", "command", "opinion", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(ctx, m.ChannelID, response)
}

// handleWhoWon handles the !who_won command
func (b *Bot) handleWhoWon(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "Analyzing the last arguments...")

	provider, args := extractProviderAndArgs(args, ai.DefaultProvider)
	model := ai.DefaultGrokModel
	persona := ai.GrokPersona
	if provider == ai.ProviderOpenAI {
		model = ai.DefaultOpenAIModel
		persona = ai.OpenAIPersona
	}

	numMessages := DefaultWhoWonMessageCount
	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			numMessages = n
		}
	}

	contextStr, err := b.formatChannelHistory(ctx, m.ChannelID, numMessages)
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to fetch channel history", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	systemMessage := fmt.Sprintf("%s\nHere are the last %d messages in this channel:\n%s\n"+
		"Based on the arguments and discussions, determine who won the arguments and why. "+
		"Be specific and fair, and explain your reasoning.", persona, numMessages, contextStr)

	response, err := b.aiClient.AskClient(ctx, "Who won the arguments in the recent conversation?",
		systemMessage, model, provider, ai.DefaultMaxTokens)
	if err != nil {
		b.logger.ErrorContext(ctx, "AI request failed", "command", "who_won", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(ctx, m.ChannelID, response)
}

// handleUserOpinion handles the !user_opinion command
func (b *Bot) handleUserOpinion(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !user_opinion @user [grok|openai] [days] [max_messages]")
		return
	}

	// Parse mentioned user
	if len(m.Mentions) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Please mention a user to analyze.")
		return
	}
	targetUser := m.Mentions[0]

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Analyzing %s...", targetUser.Username))

	provider, days, maxMessages := parseUserOpinionArgs(args)

	// Fetch messages from the user
	userMessages, err := b.fetchUserMessages(ctx, s, m.ChannelID, m.GuildID, targetUser, days, maxMessages)
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to fetch user messages", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	if len(userMessages) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No messages found for %s in the last %d days.", targetUser.Username, days))
		return
	}

	contextStr := strings.Join(userMessages, "\n")
	systemMessage := fmt.Sprintf("Here are all the messages sent by %s in the last %d days in this channel:\n%s\n",
		targetUser.Username, days, contextStr)

	response, err := b.aiClient.AskClient(ctx, fmt.Sprintf("What is your opinion of %s?", targetUser.Username),
		systemMessage, ai.DefaultOpenAIModel, provider, ai.DefaultMaxTokens)
	if err != nil {
		b.logger.ErrorContext(ctx, "AI request failed", "command", "user_opinion", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(ctx, m.ChannelID, response)
}

// parseUserOpinionArgs parses arguments for the user_opinion command
func parseUserOpinionArgs(args []string) (provider string, days int, maxMessages int) {
	provider = "openai"
	days = DefaultUserOpinionDays
	maxMessages = DefaultUserOpinionMaxMessages

	// Remove the mention from args
	argsWithoutMention := []string{}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "<@") {
			argsWithoutMention = append(argsWithoutMention, arg)
		}
	}

	provider, remainingArgs := extractProviderAndArgs(argsWithoutMention, "openai")

	if len(remainingArgs) > 0 {
		if n, err := strconv.Atoi(remainingArgs[0]); err == nil {
			days = n
			remainingArgs = remainingArgs[1:]
		}
	}

	if len(remainingArgs) > 0 {
		if n, err := strconv.Atoi(remainingArgs[0]); err == nil {
			maxMessages = n
		}
	}

	return provider, days, maxMessages
}

// fetchUserMessages fetches messages from a specific user within a time window
func (b *Bot) fetchUserMessages(ctx context.Context, s *discordgo.Session, channelID, guildID string, targetUser *discordgo.User, days int, maxMessages int) ([]string, error) {
	after := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	allMessages, err := s.ChannelMessages(channelID, maxMessages, "", "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch channel messages: %w", err)
	}

	var userMessages []string
	for _, msg := range allMessages {
		if msg.Author.ID == targetUser.ID && msg.Timestamp.After(after) {
			member, err := s.GuildMember(guildID, msg.Author.ID)
			displayName := msg.Author.Username
			if err == nil && member.Nick != "" {
				displayName = member.Nick
			}
			userMessages = append(userMessages, fmt.Sprintf("%s: %s", displayName, msg.Content))
		}
	}

	return userMessages, nil
}

// handleMost handles the !most command
func (b *Bot) handleMost(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !most [grok|openai] <question>")
		return
	}

	numMessages := DefaultMostMessageCount
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Analyzing: %s (last %d messages)...", strings.Join(args, " "), numMessages))

	provider, args := extractProviderAndArgs(args, "openai")
	question := strings.Join(args, " ")

	messages, userCounts, err := b.fetchAndCountMessages(ctx, s, m.ChannelID, m.GuildID, numMessages)
	if err != nil {
		b.logger.ErrorContext(ctx, "failed to fetch messages", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	activeUserNames := getTopActiveUsers(userCounts, TopActiveUsersCount)
	contextStr := strings.Join(messages, "\n")

	prompt := question
	if len(strings.Fields(question)) == 1 {
		prompt = fmt.Sprintf("Who is the most %s in the recent conversation?", question)
	}

	systemMessage := fmt.Sprintf("%s\nHere are the last %d messages in this channel:\n%s\n"+
		"Among the most active users (%s), answer the following question: %s. "+
		"Explain your reasoning as Coonbot.", ai.OpenAIPersona, numMessages, contextStr,
		strings.Join(activeUserNames, ", "), question)

	response, err := b.aiClient.AskClient(ctx, prompt, systemMessage, ai.DefaultOpenAIModel, provider, ai.DefaultMaxTokens)
	if err != nil {
		b.logger.ErrorContext(ctx, "AI request failed", "command", "most", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(ctx, m.ChannelID, response)
}

// fetchAndCountMessages fetches messages and counts them by user
func (b *Bot) fetchAndCountMessages(ctx context.Context, s *discordgo.Session, channelID, guildID string, numMessages int) ([]string, map[string]int, error) {
	allMessages, err := s.ChannelMessages(channelID, numMessages, "", "", "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch channel messages: %w", err)
	}

	var messages []string
	userMessageCount := make(map[string]int)

	for i := len(allMessages) - 1; i >= 0; i-- {
		msg := allMessages[i]
		if msg.Author.Bot {
			continue
		}

		member, err := s.GuildMember(guildID, msg.Author.ID)
		displayName := msg.Author.Username
		if err == nil && member.Nick != "" {
			displayName = member.Nick
		}

		messages = append(messages, fmt.Sprintf("%s: %s", displayName, msg.Content))
		userMessageCount[displayName]++
	}

	return messages, userMessageCount, nil
}

// getTopActiveUsers returns the top N most active users from a count map
func getTopActiveUsers(userCounts map[string]int, topN int) []string {
	type userCount struct {
		name  string
		count int
	}

	var counts []userCount
	for name, count := range userCounts {
		counts = append(counts, userCount{name, count})
	}

	// Simple bubble sort by count (descending)
	for i := 0; i < len(counts); i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[i].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}

	activeUserNames := []string{}
	for i := 0; i < topN && i < len(counts); i++ {
		activeUserNames = append(activeUserNames, counts[i].name)
	}

	return activeUserNames
}

// handleImageOpinion handles the !image_opinion command
func (b *Bot) handleImageOpinion(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	var imageURL string
	var customPrompt *string

	provider, args := extractProviderAndArgs(args, "openai")
	visionModel := ai.DefaultOpenAIVisionModel

	// Check for attachment first
	if len(m.Attachments) > 0 {
		imageURL = m.Attachments[0].URL
		if len(args) > 0 {
			prompt := strings.Join(args, " ")
			customPrompt = &prompt
		}
	} else if m.MessageReference != nil {
		// If replying to a message
		refMsg, err := s.ChannelMessage(m.ChannelID, m.MessageReference.MessageID)
		if err != nil {
			b.logger.ErrorContext(ctx, "failed to fetch referenced message", "error", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Could not fetch replied message: %v", err))
			return
		}
		if len(refMsg.Attachments) > 0 {
			imageURL = refMsg.Attachments[0].URL
		}
		if len(args) > 0 {
			prompt := strings.Join(args, " ")
			customPrompt = &prompt
		}
	} else if len(args) > 0 {
		// Check for image URL in args
		possibleURL := args[0]
		if strings.HasPrefix(possibleURL, "http://") || strings.HasPrefix(possibleURL, "https://") {
			imageURL = possibleURL
			if len(args) > 1 {
				prompt := strings.Join(args[1:], " ")
				customPrompt = &prompt
			}
		} else {
			prompt := strings.Join(args, " ")
			customPrompt = &prompt
		}
	}

	if imageURL == "" {
		s.ChannelMessageSend(m.ChannelID, "Please attach an image, provide a valid image URL (starting with http/https), or reply to a message with an image.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Analyzing image, one sec...")

	var opinion string
	var err error

	if provider == "grok" {
		opinion, err = b.aiClient.ImageOpinionGrok(ctx, imageURL, ai.OpenAIPersona, customPrompt)
	} else {
		b.sendThinkingMessage(ctx, s, m.ChannelID, provider, visionModel)
		opinion, err = b.aiClient.ImageOpinionOpenAI(ctx, imageURL, ai.OpenAIPersona, visionModel, ai.DefaultMaxTokens, customPrompt)
	}

	if err != nil {
		b.logger.ErrorContext(ctx, "image analysis failed", "command", "image_opinion", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error analyzing image: %v", err))
		return
	}

	b.sendLongResponse(ctx, m.ChannelID, opinion)
}

// handleRoast handles the !roast command
func (b *Bot) handleRoast(ctx context.Context, s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	var targetName string
	var roastMessage string
	var systemMessage string
	var prompt string

	// If user is mentioned
	if len(m.Mentions) > 0 {
		targetUser := m.Mentions[0]
		member, err := s.GuildMember(m.GuildID, targetUser.ID)
		targetName = targetUser.Username
		if err == nil && member.Nick != "" {
			targetName = member.Nick
		}

		systemMessage = fmt.Sprintf("%s\nRoast %s as if you were a boston comedian who grew up in the bronx. "+
			"Dont be afraid to attack sexuality, its a joke and everyone gets it. Be really really mean.", ai.OpenAIPersona, targetName)
		prompt = fmt.Sprintf("Roast %s.", targetName)
	} else if m.MessageReference != nil {
		// If command is a reply to a message
		refMsg, err := s.ChannelMessage(m.ChannelID, m.MessageReference.MessageID)
		if err != nil {
			b.logger.ErrorContext(ctx, "failed to fetch referenced message", "error", err)
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Could not fetch replied message: %v", err))
			return
		}

		member, err := s.GuildMember(m.GuildID, refMsg.Author.ID)
		targetName = refMsg.Author.Username
		if err == nil && member.Nick != "" {
			targetName = member.Nick
		}
		roastMessage = refMsg.Content

		systemMessage = fmt.Sprintf("%s\nRoast %s based on this message: '%s'. "+
			"Be a boston comedian from the bronx, don't be afraid to attack sexuality, it's a joke and everyone gets it.",
			ai.OpenAIPersona, targetName, roastMessage)
		prompt = fmt.Sprintf("Roast %s for saying: %s", targetName, roastMessage)
	} else {
		s.ChannelMessageSend(m.ChannelID, "Please mention a user or reply to a message to roast.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Cooking up a roast for %s...", targetName))

	response, err := b.aiClient.AskClient(ctx, prompt, systemMessage, ai.DefaultOpenAIModel, "openai", ai.DefaultMaxTokens)
	if err != nil {
		b.logger.ErrorContext(ctx, "AI request failed", "command", "roast", "error", err)
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(ctx, m.ChannelID, response)
}

// sendThinkingMessage sends a "thinking" message to indicate processing
func (b *Bot) sendThinkingMessage(ctx context.Context, s *discordgo.Session, channelID, provider, model string) {
	providerName := providerDisplayName(provider)
	version := getModelVersion(provider, model)
	modelLabel := model

	message := fmt.Sprintf("Thinking with %s - knowledge cutoff %s ...", modelLabel, version)
	b.logger.InfoContext(ctx, "sending thinking message",
		"channel_id", channelID,
		"provider", providerName,
		"model", model,
		"version", version)

	s.ChannelMessageSend(channelID, message)
}

// getModelVersion returns the version string for a provider and model
func getModelVersion(provider, model string) string {
	if provider == "grok" && model == ai.DefaultGrokModel {
		return ai.DefaultGrokModelVersion
	}

	if provider == "openai" && model == ai.DefaultOpenAIModel {
		return ai.DefaultOpenAIModelVersion
	}

	return ""
}
