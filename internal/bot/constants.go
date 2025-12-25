package bot

import "time"

// Discord message and command constants
const (
	MaxDiscordMessageLength       = 2000
	DefaultHistoryMessageCount    = 10
	DefaultWhoWonMessageCount     = 100
	DefaultMostMessageCount       = 100
	DefaultUserOpinionDays        = 3
	DefaultUserOpinionMaxMessages = 200
	TopActiveUsersCount           = 5
)

// Message delivery timing for human-like responses
const (
	// MinMessageDelay is the minimum delay between message chunks
	MinMessageDelay = 5000 * time.Millisecond
	// MaxMessageDelay is the maximum delay between message chunks
	MaxMessageDelay = 8000 * time.Millisecond
	// TypingSpeed simulates typing speed (milliseconds per character)
	TypingSpeed = 50 * time.Millisecond
)
