Repository onboarding for Copilot coding agents

Purpose

This document tells a coding assistant everything it needs to quickly make safe, CI-friendly changes to this repository without excessive searching. It is authoritative: follow it first and only search the repo when this file is incomplete or clearly incorrect.

Summary

- What this repo is: a Go-based Discord bot that integrates OpenAI (and optionally Grok/xAI) to answer chat commands and produce persona-driven responses. Core behavior is organized in the `internal/` directory with packages for ai, bot, config, discord, and logging.
- Language & tooling: Go module (go.mod declares go 1.24.11), standard Go toolchain (go build/go test/go vet), and uses these third-party libs: github.com/bwmarrin/discordgo, github.com/sashabaranov/go-openai, github.com/joho/godotenv.
- Size & expected complexity: small single-module repo (~15 source files in internal/). No CI workflows (.github/workflows) were found as of this file's creation.
- Testing: Unit tests exist for bot handlers, formatting utilities, and configuration validation. Run `go test ./... -v` to execute them.

Agent Coding Persona (recommended)

- Persona summary: act like a senior Go developer who has read and internalized the practical lessons from "100 Go Mistakes and How to Avoid Them" by Teiva Harsanyi. Use those lessons to prefer clear, safe, idiomatic, and testable changes.

- Key habits to enforce on every change:
  - Prefer small, minimal, well-scoped PRs with a clear description and a short test plan.
  - Always run formatting and static checks before proposing code: `go fmt ./...; go vet ./...; go mod tidy`.
  - Favor explicit error handling over panics; return errors to callers and attach context where useful.
  - Use `context.Context` for cancellations/timeouts on network or long-running operations and check it promptly.
  - Avoid global mutable state; prefer dependency injection (pass interfaces) for easier testing and mocking.
  - Write unit tests (table-driven where applicable) for logic changes; include a minimal integration/smoke test when changing runtime behavior.
  - Limit large allocations and unbounded goroutine usage; always document concurrency assumptions and cancelation behavior.
  - Keep secrets out of code and logs; never print persona strings or API keys in logs or test output.
  - Prefer the standard library for common tasks; when bringing third-party libraries, keep them minimal and well-vetted.
  - Add short comments for non-obvious decisions and follow existing repo naming and style conventions.

- When uncertain, follow this trust rule: prefer safe, conservative changes that preserve existing behavior and build/test cleanly.

Quick trust rule

- Trust the commands and file paths in this document. Only do a repo search if:
  - A command fails and the failure is not explained here
  - You need to modify files not listed here
  - The codebase has changed since this file was created

Project layout (important files)

Root files (highest priority):
- README.md — usage, commands, env var list, build hints
- go.mod, go.sum — module and dependencies (go version: 1.24.11)
- main.go — program entrypoint; calls config.LoadConfig(), bot.NewBot(), bot.Start(), waits for SIGINT

Internal packages (core implementation):
- internal/ai/
  - client.go — AI client wrappers (OpenAI/Grok) with AskClient, AskWithImage methods
  - interface.go — AI client interface definition
  - models.go — model constants and configurations
  - personas.go — persona text used by the bot (contains strong stylistic instructions; may be offensive). Do NOT leak these strings into public logs or commit secrets.
  - errors.go — AI-specific error types
- internal/bot/
  - bot.go — main bot implementation and command handlers (!ping, !ask, !opinion, !who_won, !user_opinion, !most, !image_opinion, !roast)
  - constants.go — bot-specific constants and defaults
  - formatting.go — message formatting and chunking utilities
  - formatting_test.go — unit tests for formatting functions
  - handlers_test.go — unit tests for command parsing and helper functions
- internal/config/
  - config.go — environment loader (uses github.com/joho/godotenv) and Config struct (reads DISCORD_TOKEN, OPENAI_API_KEY, XAI_API_KEY, DISCORD_POLITICS_CHANNEL)
  - config_test.go — unit tests for configuration loading and validation
  - errors.go — config-specific error types
- internal/discord/
  - session.go — Discord session wrapper and interface
- internal/logging/
  - logger.go — structured logging implementation using slog

Other files:
- .gitignore — ignores .env and build artifacts

Required environment

- The program requires environment variables (these are read by config.LoadConfig in internal/config):
  - DISCORD_TOKEN (required to run the bot)
  - OPENAI_API_KEY (required to call OpenAI, unless only using Grok)
  - XAI_API_KEY (optional; used for Grok/xAI provider)
  - DISCORD_POLITICS_CHANNEL (optional; defaults to "politics")
- The repo uses a .env file if present (godotenv.Load(".env")). The .env file is in .gitignore and should not be committed.

Validated bootstrap, build and test commands (Windows PowerShell)

I validated these commands in PowerShell in this repo. Run them from the repository root (where go.mod lives).

1) Sync and download dependencies (always do this first):
   go mod tidy
   go mod download

   Notes: these are fast for this small repo and succeeded when tested.

2) Build (production / local binary):
   go build -o discord-bot

   Alternative (platform-specific):
   $env:GOOS = 'linux'; $env:GOARCH = 'amd64'; go build -o discord-bot-linux

   Notes: building at repository root produces an executable named per the -o flag. The repository compiles with the current dependencies.

3) Run without building (for quick iteration):
   go run .

4) Vet and lint (validation):
   go vet ./...

   Optional: add a linter (not present in repo): install golangci-lint and run:
   choco install golangci-lint -y  # optional on Windows
   golangci-lint run

   Note: There is no golangci-lint config file in the repo. If you add one, include it in the PR.

5) Tests:
   go test ./... -v

   Note: Unit tests exist for internal/bot (formatting, handlers), internal/config (config validation), and pass successfully. Always run tests after making changes.

Observed outputs and validation

- go mod tidy / go mod download: completed successfully on a local run.
- go build ./...: completed successfully when executed from the repo root.
- go vet ./...: no notable output (no vet issues found during validation).
- go test ./... -v: all tests pass (bot formatting, handlers, and config validation tests).

Common gotchas / repo-specific notes

- .env handling: code calls godotenv.Load(".env"). If you run the program locally, create a .env file in the repo root with DISCORD_TOKEN and OPENAI_API_KEY. Do NOT commit .env.
- Persona content: `internal/ai/personas.go` contains aggressive, explicit persona strings. Be cautious: do not introduce changes that publish those strings to public logs or telemetry. Don't accidentally include them in test output or PR descriptions.
- No CI workflows detected: assume no automated checks run on push. Before opening a PR, run the bootstrap/build/test/lint steps above locally and ensure they pass.
- Secrets: never commit tokens/keys. The repo already .gitignores .env.
- Package organization: All core logic is in `internal/` packages. The main.go file is minimal and only wires everything together.

Change guidance (where to edit for common tasks)

- Add a new bot command: edit `internal/bot/bot.go` and add a new case in the messageHandler switch and implement the handler below. Add tests in `internal/bot/handlers_test.go`.
- Change config/env keys: edit `internal/config/config.go` and the LoadConfig function. Add validation in Validate(). Update tests in `internal/config/config_test.go`.
- Change AI behavior: edit `internal/ai/client.go` and `internal/ai/personas.go` (personas.go only for persona text — handle carefully).
- Add tests: create `_test.go` files alongside the package files and use `go test ./... -v`.
- Change logging: edit `internal/logging/logger.go` for structured logging modifications.

Safety & PR checklist (small, concrete)

Before pushing a PR, run and confirm these locally in PowerShell:
- go mod tidy; go mod download
- go vet ./...
- go build -o discord-bot
- go test ./... -v
- Verify no secrets were added and .env was not committed
- Search for TODO/HACK strings only if you need context: `Select-String -Pattern "TODO|HACK" -Path * -SimpleMatch`

If any command fails

- If `go mod tidy` fails: check go.mod for malformed entries. Try `go mod edit` or remove problem require lines.
- If `go build` fails: run `go vet ./...` to catch obvious issues; inspect the build error line and open the referenced file.
- If `go test` fails after adding tests: run them with `-run` to isolate failing tests.

When to search the repo
`
- If an instruction in this file points to a file that no longer exists or the build command`s fail in a way not described here, then search the codebase for changed file names, CI workflows, or new scripts.

Contact notes

- This file was generated by an onboarding pass on the repository. If you find anything out-of-date, update `.github/copilot-instructions.md` in the repo root.

End of file
