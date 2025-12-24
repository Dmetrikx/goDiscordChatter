package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Load configuration
	config := LoadConfig()

	if config.DiscordToken == "" {
		log.Fatal("DISCORD_TOKEN environment variable is not set")
	}

	if config.XAIAPIKey == "" {
		log.Fatal("XAI_API_KEY environment variable is not set")
	}

	// Create bot
	bot, err := NewBot(config)
	if err != nil {
		log.Fatalf("Error creating bot: %v", err)
	}

	// Start bot
	if err := bot.Start(); err != nil {
		log.Fatalf("Error starting bot: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown
	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close the bot
	log.Println("Shutting down bot...")
	bot.Close()
}
