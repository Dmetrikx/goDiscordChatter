Repository onboarding for Copilot coding agents

Purpose

This document tells a coding assistant everything it needs to quickly make safe, CI-friendly changes to this repository without excessive searching. It is authoritative: follow it first and only search the repo when this file is incomplete or clearly incorrect.

Summary

- What this repo is: a Go-based Discord bot that integrates OpenAI (and optionally Grok/xAI) to answer chat commands and produce persona-driven responses. Core behavior lives in Go source files at the repository root (not in a nested "go/" folder).
- Language & tooling: Go module (go.mod declares go 1.24.11), standard Go toolchain (go build/go test/go vet), and uses these third-party libs: github.com/bwmarrin/discordgo, github.com/sashabaranov/go-openai, github.com/joho/godotenv.
- Size & expected complexity: small single-module repo (~10 source files). No CI workflows (.github/workflows) were found as of this file's creation.

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
- main.go — program entrypoint; calls LoadConfig(), NewBot(), Start(), waits for SIGINT
- bot.go — main bot implementation and command handlers (!ping, !ask, !opinion, !who_won, !user_opinion, !most, !image_opinion, !roast)
- client.go — AI client wrappers (OpenAI/Grok) used by bot (refer to for model/HTTP code)
- config.go — environment loader (uses github.com/joho/godotenv) and Config struct (reads DISCORD_TOKEN, OPENAI_API_KEY, XAI_API_KEY, DISCORD_POLITICS_CHANNEL)
- constants.go — runtime constants and defaults
- personas.go — persona text used by the bot (contains strong stylistic instructions; may be offensive). Do NOT leak these strings into public logs or commit secrets.
- .gitignore — ignores .env and build artifacts

Notes about README vs reality

- README contains a "go/" prefix in the Project Structure example, but actual source files are at the repo root. Use the repository root as working directory when running build/test commands.

Required environment

- The program requires environment variables (these are read by LoadConfig):
  - DISCORD_TOKEN (required to run the bot)
  - OPENAI_API_KEY (required to call OpenAI)
  - XAI_API_KEY (optional; used for Grok/xAI provider)
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

   Note: there are currently no test files; `go test` exits with "[no test files]" for this module. If you add tests, run these commands and ensure they pass locally.

Observed outputs and validation

- go mod tidy / go mod download: completed successfully on a local run.
- go build ./...: completed successfully when executed from the repo root.
- go vet ./...: no notable output (no vet issues found during validation).
- go test ./... -v: returned `[no test files]` (no tests present).

Common gotchas / repo-specific notes

- .env handling: code calls godotenv.Load(".env"). If you run the program locally, create a .env file in the repo root with DISCORD_TOKEN and OPENAI_API_KEY. Do NOT commit .env.
- README mismatch: README references a "go/" subdirectory for examples; the actual code is in the root. Use repository root.
- Persona content: `personas.go` contains aggressive, explicit persona strings. Be cautious: do not introduce changes that publish those strings to public logs or telemetry. Don't accidentally include them in test output or PR descriptions.
- No CI workflows detected: assume no automated checks run on push. Before opening a PR, run the bootstrap/build/test/lint steps above locally and ensure they pass.
- Secrets: never commit tokens/keys. The repo already .gitignores .env.

Change guidance (where to edit for common tasks)

- Add a new bot command: edit `bot.go` and add a new case in the messageHandler switch and implement the handler below.
- Change config/env keys: edit `config.go` and the LoadConfig function. Keep .env handling.
- Change AI behavior: edit `client.go` and `personas.go` (personas.go only for persona text — handle carefully).
- Add tests: create `_test.go` files alongside the package files and use `go test ./... -v`.

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
