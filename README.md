# goDiscordChatter

A small Go-based Discord bot that integrates OpenAI (and optionally Grok/xAI) to answer chat commands and produce persona-driven responses.

## Status
Prototype / personal tool. Use for local experimentation; do not run with production keys without review.

## Features

- OpenAI integration for intelligent, persona-driven responses
- Optional Grok (xAI) integration as an alternative AI provider
- Command handler for chat interaction, argument analysis, and image opinions
- Modular structure for easy feature expansion

## Commands & Usage

Interact with the bot using the following commands in any Discord channel where the bot is present:

### `!ping`
Check if the bot is online.
- Example: `!ping`

### `!ask <question>`
Ask the bot any question and get a persona-driven response.
- Example: `!ask What do you think about Boston politics?`
- Example: `!ask Who would win in a fight, Batman or Superman?`

### `!opinion [num_messages]`
Get the bot's opinion or summary on the last few messages in the channel.
- Example: `!opinion` (default: 10 messages)
- Example: `!opinion 20` (analyzes last 20 messages)

### `!who_won [num_messages]`
Analyze recent arguments and determine who won.
- Example: `!who_won` (default: 100 messages)
- Example: `!who_won 50` (analyzes last 50 messages)

### `!user_opinion <@user> [days] [max_messages]`
Get the bot's opinion on a specific user based on their recent messages.
- Example: `!user_opinion @Alice` (default: 3 days, 200 messages)
- Example: `!user_opinion @Bob 5 100` (analyzes Bob's last 100 messages over 5 days)

### `!most <question>`
Ask who is the most X or most likely to do Y in the chat.
- Example: `!most helpful` (default: last 100 messages)
- Example: `!most Who is most likely to start an argument?`

### `!image_opinion <image_url> [custom_prompt]` / attach an image / reply to an image
Form an opinion on an image by:
- Attaching an image and typing `!image_opinion` (optionally add a custom prompt after the command)
  - Example: `!image_opinion Give a funny take on this picture.`
- Providing an image URL: `!image_opinion https://example.com/image.jpg` (optionally add a custom prompt after the URL)
  - Example: `!image_opinion https://example.com/image.jpg What do you think of this meme?`
- Replying to a message with an image attachment and typing `!image_opinion` (optionally add a custom prompt after the command)
  - Example: *(reply to an image)* `!image_opinion Be controversial about this photo.`

### `!roast <@user>` or reply to a message
Roast a user in a witty, funny, and lighthearted way. You can either mention a user or reply to their message.
- Example: `!roast @Alice`
- Example: *(reply to a message)* `!roast`

### Provider Override (OpenAI/Grok)
You can override the AI provider for any command that uses language models by prefixing your prompt with `grok` or `openai`:
- Example: `!ask grok Who are you?` (uses Grok)
- Example: `!ask openai Who are you?` (uses OpenAI)
- Example: `!most grok Who is most likely to start an argument?`
- Example: `!opinion grok` (uses Grok for opinion)
- Example: `!who_won grok` (uses Grok for argument analysis)
- Example: `!user_opinion @Alice grok` (uses Grok for user analysis)
- Example: `!image_opinion grok https://example.com/image.jpg` (uses Grok for image analysis)

If no provider is specified, OpenAI is used by default.

## Setup

### Prerequisites
- Go 1.24.11 (as declared in `go.mod`)
- Discord Bot Token
- OpenAI API Key
- (Optional) xAI API Key for Grok support

### Installation

1. Clone the repository and navigate to the repository root (where `go.mod` is located).

2. Create a `.env` file in the repository root with the required environment variables:
   
   Required:
   ```
   DISCORD_TOKEN=your_discord_bot_token_here
   OPENAI_API_KEY=your_openai_api_key_here
   ```
   
   Optional:
   ```
   XAI_API_KEY=your_xai_api_key_here
   DISCORD_POLITICS_CHANNEL=politics_channel_id_here
   ```

   **IMPORTANT**: Do NOT commit the `.env` file. It is already in `.gitignore`.

3. Install and sync dependencies (PowerShell):
   ```powershell
   go mod tidy
   go mod download
   ```

4. Build the bot:
   ```powershell
   go build -o discord-bot
   ```

5. Run the bot:
   ```powershell
   .\discord-bot
   ```

   Or run directly without building:
   ```powershell
   go run .
   ```

## Project Structure

Repository root (where `go.mod` lives):
```
├── main.go           - Entry point, initializes and starts the bot
├── bot.go            - Discord bot logic and command handlers
├── client.go         - OpenAI and Grok API client implementations
├── config.go         - Configuration management
├── constants.go      - Application constants
├── personas.go       - Bot persona definitions (sensitive content)
├── go.mod            - Go module definition
├── go.sum            - Go module checksums
├── .env              - Local environment variables (gitignored, do not commit)
├── .gitignore        - Git ignore rules
└── README.md         - This file
```

## Adding Features

Add new commands in `bot.go` by:
1. Adding a new case in the `messageHandler` switch statement
2. Implementing the command handler function
3. Running validation checks before committing (see below)

## Validation & Testing

Before committing changes, run these checks locally (PowerShell):

```powershell
# Format code
go fmt ./...

# Static analysis
go vet ./...

# Tidy dependencies
go mod tidy

# Build
go build -o discord-bot

# Run tests (currently no tests present)
go test ./... -v
```

## Building for Different Platforms

Build for Linux:
```powershell
$env:GOOS = 'linux'; $env:GOARCH = 'amd64'; go build -o discord-bot-linux
```

Build for Windows:
```powershell
$env:GOOS = 'windows'; $env:GOARCH = 'amd64'; go build -o discord-bot.exe
```

Build for macOS:
```powershell
$env:GOOS = 'darwin'; $env:GOARCH = 'amd64'; go build -o discord-bot-macos
```

## Dependencies

- [discordgo](https://github.com/bwmarrin/discordgo) - Discord API wrapper
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI API client
- [godotenv](https://github.com/joho/godotenv) - Environment variable loader

## Security & Safety

- **Never commit `.env`** - The `.env` file is gitignored for security. It contains sensitive tokens and API keys.
- **Persona content warning** - `personas.go` contains strong persona strings. Do not log or publish these strings in test output, commits, or public logs.
- **API key safety** - Never print API keys or tokens in logs or console output.
- **Limited scope** - This is a personal/prototype bot. Review and test thoroughly before using with production credentials.

## Notes

- The bot uses the persona defined in `personas.go` to generate responses.
- Grok support requires the `XAI_API_KEY` environment variable to be set.
- Default AI provider is OpenAI; use provider override syntax to switch to Grok per-command.
- Run all validation checks before committing changes (see "Validation & Testing" section).

## Troubleshooting

### Bot doesn't respond
- Ensure the bot has proper permissions in your Discord server
- Check that Message Content Intent is enabled in the Discord Developer Portal
- Verify your `DISCORD_TOKEN` is correct

### OpenAI API errors
- Check that your `OPENAI_API_KEY` is valid and has sufficient credits
- Ensure you're using a model you have access to

### Grok API errors
- Verify your `XAI_API_KEY` is set correctly
- Ensure you have access to Grok API
