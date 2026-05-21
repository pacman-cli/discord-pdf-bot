# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Discord bot in Go. Watches `./pdfs/` folder, registers each PDF as a Discord slash command. Uses SQLite for metadata storage, fsnotify for hot-reload. Currently mid-refactor from single-file to clean architecture — see "Incomplete Work" below.

## Commands

```bash
# Run
go run cmd/bot/main.go

# Build
go build -o discord-pdf-bot ./cmd/bot/

# Test
go test ./... -v

# Run single test
go test ./internal/adapter/repository/ -v -run TestSearch

# Dependencies
go mod tidy
```

## Architecture

Clean architecture with 4 layers. Entry point is `cmd/bot/main.go` (currently a minimal stub — full DI wiring not yet implemented).

**Domain** (`internal/domain/`) — zero external dependencies:
- `entity/pdf.go` — PDF struct (ID, Name, Filename, Path, Description, CategoryID, UploadedBy, PageCount, FileSize, timestamps)
- `port/pdf_repository.go` — `PDFRepository` interface (GetByName, GetAll, GetByCategory, Search, Create, Update, Delete)
- `port/storage.go` — `StoragePort` interface (Save, Delete, Read, List)
- `errors.go` — domain errors (ErrPDFNotFound, ErrDuplicateName, ErrPermissionDenied, ErrInvalidInput)

**Use Case** (`internal/usecase/`) — application services:
- `pdf_service.go` — `PDFService` wrapping PDFRepository. Includes `SyncFromDisk()` for two-way sync (add new, remove deleted).

**Adapter** (`internal/adapter/`):
- `repository/sqlite_pdf.go` — SQLite implementation of PDFRepository
- `storage/disk_storage.go` — disk implementation of StoragePort

**Infrastructure** (`internal/infrastructure/`):
- `database/sqlite.go` — SQLite connection, migrations (creates `categories`, `pdfs`, `permissions` tables with indexes)
- `watcher/fsnotify.go` — filesystem watcher with 500ms debounce

## Key Dependencies

- `discordgo` v0.29.0 — Discord API
- `fsnotify` v1.9.0 — filesystem watcher
- `modernc.org/sqlite` v1.50.1 — pure-Go SQLite driver

## Config

Environment variables (not yet wired in main.go):
- `DISCORD_BOT_TOKEN` — Discord bot token
- `GUILD_ID` — Discord server ID
- `ADMIN_ROLE` — role name for admin commands (default: "PDF Admin")

Files:
- `.env` — environment variables (gitignored)
- `data/bot.db` — SQLite database (gitignored)
- `pdfs/` — PDF storage directory

## Database Schema

Three tables created by `database.Migrate()`:
- `categories` (id, name, description, created_at)
- `pdfs` (id, name, filename, path, description, category_id FK, uploaded_by, page_count, file_size, timestamps)
- `permissions` (id, pdf_id FK, category_id FK, role_id, user_id, allowed, created_at)

Indexes on: `pdfs(name)`, `pdfs(category_id)`, `permissions(pdf_id)`, `permissions(category_id)`. Default category seeded.

## Incomplete Work

See `docs/superpowers/plans/2026-05-18-pdf-bot-upgrade.md` for the full implementation plan. Tasks 1-6 are done (directory structure, SQLite, watcher, repository, storage, PDF service). Remaining:

- **Discord adapter** (`internal/adapter/discord/`) — bot setup, interaction handlers, embed builders. Not yet created.
- **Category service** — entity, port, repository, service. Not yet created.
- **Permission service** — entity, port, repository, service. Not yet created.
- **Main.go DI wiring** — currently a stub; needs to wire database, repositories, services, bot, and file watcher.
- **Admin commands** — upload, delete, pdf info/edit.
- **Task 14** — update this CLAUDE.md (this task).

The `main.go` entry point is a minimal placeholder. The `cmd/bot/main.go` in the plan (Task 11) shows the intended full wiring.
