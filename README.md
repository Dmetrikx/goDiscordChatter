# Discord OpenAI Bot (Go Version)

A Go implementation of the Discord bot using DiscordGo and OpenAI/Grok APIs.

## Features

- Modular structure for easy feature expansion
- OpenAI integration for intelligent, persona-driven responses
- Grok (xAI) integration as an alternative AI provider
- Simple command handler (e.g., `!ping`, `!ask`, `!opinion`, `!who_won`)

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
- Go 1.21 or higher
- Discord Bot Token
- OpenAI API Key
- (Optional) xAI API Key for Grok support

### Installation

1. Copy your environment variables to a `.env` file in the root directory (see `.env.example` if available).
   
   Required environment variables:
   ```
   DISCORD_TOKEN=your_discord_bot_token_here
   OPENAI_API_KEY=your_openai_api_key_here
   ```
   
   Optional environment variables:
   ```
   XAI_API_KEY=your_xai_api_key_here
   DISCORD_POLITICS_CHANNEL=politics
   ```

2. Install dependencies:
   ```bash
   cd go
   go mod download
   ```

3. Build the bot:
   ```bash
   go build -o discord-bot
   ```

4. Run the bot:
   ```bash
   ./discord-bot
   ```

   Or run directly without building:
   ```bash
   go run .
   ```

## Project Structure

```
go/
├── main.go       - Entry point, initializes and starts the bot
├── bot.go        - Discord bot logic and command handlers
├── client.go     - OpenAI and Grok API client implementations
├── config.go     - Configuration management
├── constants.go  - Application constants
├── personas.go   - Bot persona definitions
├── go.mod        - Go module definition
└── go.sum        - Go module checksums
```

## Adding Features

Add new commands in the `bot.go` file by:
1. Adding a new case in the `messageHandler` switch statement
2. Implementing the command handler function

## Dependencies

- [discordgo](https://github.com/bwmarrin/discordgo) - Discord API wrapper
- [go-openai](https://github.com/sashabaranov/go-openai) - OpenAI API client
- [godotenv](https://github.com/joho/godotenv) - Environment variable loader

## Notes

- The `.env` file is gitignored for security.
- The bot uses the persona defined in `personas.go` to respond as "Coonbot," a political raccoon from Boston.
- Grok support requires the `XAI_API_KEY` environment variable to be set.

## Building for Different Platforms

Build for Linux:
```bash
GOOS=linux GOARCH=amd64 go build -o discord-bot-linux
```

Build for Windows:
```bash
GOOS=windows GOARCH=amd64 go build -o discord-bot.exe
```

Build for macOS:
```bash
GOOS=darwin GOARCH=amd64 go build -o discord-bot-macos
```

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
