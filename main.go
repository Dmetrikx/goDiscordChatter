package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dmetrikx/goDiscordChatter/internal/bot"
	"github.com/Dmetrikx/goDiscordChatter/internal/config"
	"github.com/Dmetrikx/goDiscordChatter/internal/logging"
)

func main() {
	// Create logger
	logger := logging.NewTextLogger()
	ctx := context.Background()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.ErrorContext(ctx, "failed to load configuration", "error", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		logger.ErrorContext(ctx, "invalid configuration", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "configuration loaded successfully")

	// Create bot
	bot, err := bot.NewBot(cfg, logger)
	if err != nil {
		logger.ErrorContext(ctx, "failed to create bot", "error", err)
		os.Exit(1)
	}

	// Start bot
	if err := bot.Start(ctx); err != nil {
		logger.ErrorContext(ctx, "failed to start bot", "error", err)
		os.Exit(1)
	}

	// Wait for interrupt signal to gracefully shutdown
	logger.InfoContext(ctx, "bot is now running", "message", "Press CTRL-C to exit")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close the bot with timeout
	logger.InfoContext(ctx, "shutting down bot...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := bot.Close(shutdownCtx); err != nil {
		logger.ErrorContext(shutdownCtx, "error during shutdown", "error", err)
		os.Exit(1)
	}

	logger.InfoContext(ctx, "bot shutdown complete")
}
