# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Discord bot in Go. Watches `./pdfs/` folder, registers each PDF as a Discord slash command (filename minus `.pdf` = command name). Caches PDF bytes in memory. Uses fsnotify for hot-reload on folder changes. Responds to slash commands by uploading PDF as file attachment.

## Commands

```bash
# Run
go run main.go

# Build
go build -o discord-pdf-bot .

# Dependencies
go mod tidy
```

No tests, linter, CI/CD, or Docker configured.

## Architecture

Single-file app: `main.go` (208 lines). No packages or subdirectories.

**Dependencies:** `discordgo` (Discord API), `fsnotify` (filesystem watcher)

**Startup flow:**
1. Read `DISCORD_BOT_TOKEN` from env
2. Create Discord session, register `interactionCreate` handler
3. `scanPDFs()` → `loadPDFsToCache()` → `syncCommands()`
4. `watchPDFs()` goroutine for live folder monitoring
5. Block on OS signal (SIGINT/SIGTERM)

**Key globals:** `pdfCache` (map[string][]byte — file contents), `pdfFiles` (map[string]string — command name → path)

**Command sync:** Two-phase — deletes stale commands, creates new ones. Each command description: `"Get the {name} PDF"`

**Naming convention:** PDF filenames are sanitized for Discord command names (see commit `9d3008b`).

## Config

- `.env` — contains `DISCORD_BOT_TOKEN` (gitignored)
- `pdfs/` — PDF files directory (contents = slash commands)
