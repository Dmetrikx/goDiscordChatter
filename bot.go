package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Bot represents the Discord bot
type Bot struct {
	session  *discordgo.Session
	aiClient *AIClient
	config   *Config
}

// NewBot creates a new bot instance
func NewBot(config *Config) (*Bot, error) {
	session, err := discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsMessageContent

	aiClient := NewAIClient(config.OpenAIAPIKey, config.XAIAPIKey)

	bot := &Bot{
		session:  session,
		aiClient: aiClient,
		config:   config,
	}

	// Register message handler
	session.AddHandler(bot.messageHandler)

	return bot, nil
}

// Start starts the bot
func (b *Bot) Start() error {
	err := b.session.Open()
	if err != nil {
		return fmt.Errorf("error opening connection: %w", err)
	}

	user, err := b.session.User("@me")
	if err != nil {
		return fmt.Errorf("error obtaining account details: %w", err)
	}

	log.Printf("Logged in as %s (ID: %s)", user.Username, user.ID)
	log.Println("------")
	return nil
}

// Close closes the bot session
func (b *Bot) Close() {
	b.session.Close()
}

// messageHandler handles incoming messages
func (b *Bot) messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
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

	// Route to appropriate command handler
	switch command {
	case "ping":
		b.handlePing(s, m)
	case "ask":
		b.handleAsk(s, m, args)
	case "opinion":
		b.handleOpinion(s, m, args)
	case "who_won":
		b.handleWhoWon(s, m, args)
	case "user_opinion":
		b.handleUserOpinion(s, m, args)
	case "most":
		b.handleMost(s, m, args)
	case "image_opinion":
		b.handleImageOpinion(s, m, args)
	case "roast":
		b.handleRoast(s, m, args)
	}
}

// handlePing responds with "Pong!"
func (b *Bot) handlePing(s *discordgo.Session, m *discordgo.MessageCreate) {
	s.ChannelMessageSend(m.ChannelID, "Pong!")
}

// extractProviderAndArgs extracts provider from arguments
func extractProviderAndArgs(args []string, defaultProvider string) (string, []string) {
	provider := defaultProvider
	if len(args) > 0 && (strings.ToLower(args[0]) == "grok" || strings.ToLower(args[0]) == "openai") {
		provider = strings.ToLower(args[0])
		args = args[1:]
	}
	return provider, args
}

// sendLongResponse sends long responses in 2000-char chunks
func (b *Bot) sendLongResponse(s *discordgo.Session, channelID, response string) {
	for i := 0; i < len(response); i += 2000 {
		end := i + 2000
		if end > len(response) {
			end = len(response)
		}
		s.ChannelMessageSend(channelID, response[i:end])
	}
}

// handleAsk handles the !ask command
func (b *Bot) handleAsk(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !ask [grok|openai] <question>")
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Thinking...")

	provider, args := extractProviderAndArgs(args, "openai")
	prompt := strings.Join(args, " ")

	response, err := b.aiClient.AskClient(prompt, OpenAIPersona, DefaultOpenAIModel, provider, DefaultMaxTokens)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(s, m.ChannelID, response)
}

// formatChannelHistory fetches and formats recent messages
func (b *Bot) formatChannelHistory(s *discordgo.Session, channelID string, numMessages int) (string, error) {
	messages, err := s.ChannelMessages(channelID, numMessages, "", "", "")
	if err != nil {
		return "", err
	}

	// Reverse the messages to show oldest first
	var formatted []string
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		member, err := s.GuildMember(msg.GuildID, msg.Author.ID)
		displayName := msg.Author.Username
		if err == nil && member.Nick != "" {
			displayName = member.Nick
		}
		formatted = append(formatted, fmt.Sprintf("%s: %s", displayName, msg.Content))
	}

	return strings.Join(formatted, "\n"), nil
}

// handleOpinion handles the !opinion command
func (b *Bot) handleOpinion(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "Let me think about what everyone has been saying...")

	provider, args := extractProviderAndArgs(args, "openai")
	numMessages := 10

	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			numMessages = n
		}
	}

	contextStr, err := b.formatChannelHistory(s, m.ChannelID, numMessages)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	systemMessage := fmt.Sprintf("%s\nHere are the last %d messages in this channel:\n%s\n"+
		"Form an opinion or summary about the conversation.", OpenAIPersona, numMessages, contextStr)

	response, err := b.aiClient.AskClient("What is your opinion on the recent conversation?",
		systemMessage, DefaultOpenAIModel, provider, DefaultMaxTokens)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(s, m.ChannelID, response)
}

// handleWhoWon handles the !who_won command
func (b *Bot) handleWhoWon(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	s.ChannelMessageSend(m.ChannelID, "Analyzing the last arguments...")

	provider, args := extractProviderAndArgs(args, "openai")
	numMessages := 100

	if len(args) > 0 {
		if n, err := strconv.Atoi(args[0]); err == nil {
			numMessages = n
		}
	}

	contextStr, err := b.formatChannelHistory(s, m.ChannelID, numMessages)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	systemMessage := fmt.Sprintf("%s\nHere are the last %d messages in this channel:\n%s\n"+
		"Based on the arguments and discussions, determine who won the arguments and why. "+
		"Be specific and fair, and explain your reasoning.", OpenAIPersona, numMessages, contextStr)

	response, err := b.aiClient.AskClient("Who won the arguments in the recent conversation?",
		systemMessage, DefaultOpenAIModel, provider, DefaultMaxTokens)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(s, m.ChannelID, response)
}

// handleUserOpinion handles the !user_opinion command
func (b *Bot) handleUserOpinion(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
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

	// Remove the mention from args
	argsWithoutMention := []string{}
	for _, arg := range args {
		if !strings.HasPrefix(arg, "<@") {
			argsWithoutMention = append(argsWithoutMention, arg)
		}
	}

	provider, remainingArgs := extractProviderAndArgs(argsWithoutMention, "openai")
	days := 3
	maxMessages := 200

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

	// Fetch messages from the user
	after := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	allMessages, err := s.ChannelMessages(m.ChannelID, maxMessages, "", "", "")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	var userMessages []string
	for _, msg := range allMessages {
		if msg.Author.ID == targetUser.ID && msg.Timestamp.After(after) {
			member, err := s.GuildMember(msg.GuildID, msg.Author.ID)
			displayName := msg.Author.Username
			if err == nil && member.Nick != "" {
				displayName = member.Nick
			}
			userMessages = append(userMessages, fmt.Sprintf("%s: %s", displayName, msg.Content))
		}
	}

	if len(userMessages) == 0 {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("No messages found for %s in the last %d days.", targetUser.Username, days))
		return
	}

	contextStr := strings.Join(userMessages, "\n")
	systemMessage := fmt.Sprintf("Here are all the messages sent by %s in the last %d days in this channel:\n%s\n",
		targetUser.Username, days, contextStr)

	response, err := b.aiClient.AskClient(fmt.Sprintf("What is your opinion of %s?", targetUser.Username),
		systemMessage, DefaultOpenAIModel, provider, DefaultMaxTokens)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(s, m.ChannelID, response)
}

// handleMost handles the !most command
func (b *Bot) handleMost(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	if len(args) == 0 {
		s.ChannelMessageSend(m.ChannelID, "Usage: !most [grok|openai] <question>")
		return
	}

	numMessages := 100
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Analyzing: %s (last %d messages)...", strings.Join(args, " "), numMessages))

	provider, args := extractProviderAndArgs(args, "openai")
	question := strings.Join(args, " ")

	// Fetch messages and count by user
	allMessages, err := s.ChannelMessages(m.ChannelID, numMessages, "", "", "")
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error fetching messages: %v", err))
		return
	}

	var messages []string
	userMessageCount := make(map[string]int)

	for i := len(allMessages) - 1; i >= 0; i-- {
		msg := allMessages[i]
		if msg.Author.Bot {
			continue
		}

		member, err := s.GuildMember(msg.GuildID, msg.Author.ID)
		displayName := msg.Author.Username
		if err == nil && member.Nick != "" {
			displayName = member.Nick
		}

		messages = append(messages, fmt.Sprintf("%s: %s", displayName, msg.Content))
		userMessageCount[displayName]++
	}

	// Get top 5 active users
	type userCount struct {
		name  string
		count int
	}
	var counts []userCount
	for name, count := range userMessageCount {
		counts = append(counts, userCount{name, count})
	}

	// Simple sort by count
	for i := 0; i < len(counts); i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[i].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}

	activeUserNames := []string{}
	for i := 0; i < 5 && i < len(counts); i++ {
		activeUserNames = append(activeUserNames, counts[i].name)
	}

	contextStr := strings.Join(messages, "\n")

	prompt := question
	if len(strings.Fields(question)) == 1 {
		prompt = fmt.Sprintf("Who is the most %s in the recent conversation?", question)
	}

	systemMessage := fmt.Sprintf("%s\nHere are the last %d messages in this channel:\n%s\n"+
		"Among the most active users (%s), answer the following question: %s. "+
		"Explain your reasoning as Coonbot.", OpenAIPersona, numMessages, contextStr,
		strings.Join(activeUserNames, ", "), question)

	response, err := b.aiClient.AskClient(prompt, systemMessage, DefaultOpenAIModel, provider, DefaultMaxTokens)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(s, m.ChannelID, response)
}

// handleImageOpinion handles the !image_opinion command
func (b *Bot) handleImageOpinion(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
	var imageURL string
	var customPrompt *string

	provider, args := extractProviderAndArgs(args, "openai")

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
		opinion, err = b.aiClient.ImageOpinionGrok(imageURL, OpenAIPersona, customPrompt)
	} else {
		opinion, err = b.aiClient.ImageOpinionOpenAI(imageURL, OpenAIPersona, DefaultOpenAIVisionModel, DefaultMaxTokens, customPrompt)
	}

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error analyzing image: %v", err))
		return
	}

	b.sendLongResponse(s, m.ChannelID, opinion)
}

// handleRoast handles the !roast command
func (b *Bot) handleRoast(s *discordgo.Session, m *discordgo.MessageCreate, args []string) {
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
			"Dont be afraid to attack sexuality, its a joke and everyone gets it. Be really really mean.", OpenAIPersona, targetName)
		prompt = fmt.Sprintf("Roast %s.", targetName)
	} else if m.MessageReference != nil {
		// If command is a reply to a message
		refMsg, err := s.ChannelMessage(m.ChannelID, m.MessageReference.MessageID)
		if err != nil {
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
			OpenAIPersona, targetName, roastMessage)
		prompt = fmt.Sprintf("Roast %s for saying: %s", targetName, roastMessage)
	} else {
		s.ChannelMessageSend(m.ChannelID, "Please mention a user or reply to a message to roast.")
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Cooking up a roast for %s...", targetName))

	response, err := b.aiClient.AskClient(prompt, systemMessage, DefaultOpenAIModel, "openai", DefaultMaxTokens)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Error: %v", err))
		return
	}

	b.sendLongResponse(s, m.ChannelID, response)
}
