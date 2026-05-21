# Discord PDF Bot Upgrade Design

## Overview

Upgrade the discord-pdf-bot from a minimal single-file bot to a full-featured PDF management system with clean architecture.

**Current state:** 208-line `main.go`, globals, in-memory cache, no DB, no tests.

**Target state:** Clean architecture, SQLite storage, 7 new features, proper error handling.

## Approach

**Architecture First, Features Incremental** — refactor to clean architecture first, then add features one by one.

## Architecture

### Layers

```
┌─────────────────────────────────────────┐
│           Adapter Layer                 │
│  (Discord handlers, filesystem, HTTP)   │
├─────────────────────────────────────────┤
│           Use Case Layer                │
│  (Application services, orchestration)  │
├─────────────────────────────────────────┤
│           Domain Layer                  │
│  (Entities, ports/interfaces, rules)    │
├─────────────────────────────────────────┤
│           Infrastructure Layer          │
│  (SQLite, fsnotify, discordgo)          │
└─────────────────────────────────────────┘
```

**Domain** — pure Go, no dependencies. Entities: `PDF`, `Category`, `User`, `Permission`. Ports: `PDFRepository`, `StoragePort`, `DiscordPort`.

**Use Case** — application logic. Services: `PDFService`, `SearchService`, `CategoryService`, `PermissionService`, `AdminService`. Calls domain ports.

**Adapter** — implements ports. `SQLitePDFRepo`, `DiskStorage`, `DiscordBot`. Translates external ↔ domain.

**Infrastructure** — concrete implementations. `discordgo` session, `fsnotify` watcher, SQLite connection.

Dependency rule: outer depends on inner, never reverse.

## Package Structure

```
discord-pdf-bot/
├── cmd/
│   └── bot/
│       └── main.go              # Entry point, DI wiring
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── pdf.go           # PDF entity
│   │   │   ├── category.go      # Category entity
│   │   │   ├── user.go          # User entity
│   │   │   └── permission.go    # Permission entity
│   │   └── port/
│   │       ├── pdf_repository.go    # PDF repo interface
│   │       ├── storage.go           # Storage interface
│   │       └── discord.go           # Discord interface
│   ├── usecase/
│   │   ├── pdf_service.go       # CRUD, search, list
│   │   ├── category_service.go  # Category management
│   │   ├── permission_service.go # Role/permission checks
│   │   └── admin_service.go     # Upload, delete, metadata
│   ├── adapter/
│   │   ├── discord/
│   │   │   ├── bot.go           # Bot setup, command registration
│   │   │   ├── handlers.go      # Interaction handlers
│   │   │   └── embeds.go        # Rich embed builders
│   │   ├── repository/
│   │   │   └── sqlite_pdf.go    # SQLite implementation
│   │   └── storage/
│   │       └── disk_storage.go  # File read/write
│   └── infrastructure/
│       ├── database/
│       │   └── sqlite.go        # Connection, migrations
│       └── watcher/
│           └── fsnotify.go      # Filesystem watcher
├── pdfs/                        # PDF storage
├── data/
│   └── bot.db                   # SQLite database
├── go.mod
├── go.sum
└── CLAUDE.md
```

- `internal/` — not importable externally
- `cmd/bot/main.go` — only entry point, wires dependencies
- Domain has zero external dependencies
- Each adapter implements one port

## Database Schema

### PDFs Table
```sql
CREATE TABLE pdfs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    filename    TEXT NOT NULL,
    path        TEXT NOT NULL,
    description TEXT DEFAULT '',
    category_id INTEGER,
    uploaded_by TEXT,
    page_count  INTEGER DEFAULT 0,
    file_size   INTEGER DEFAULT 0,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

### Categories Table
```sql
CREATE TABLE categories (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

### Permissions Table
```sql
CREATE TABLE permissions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    pdf_id      INTEGER,
    category_id INTEGER,
    role_id     TEXT,
    user_id     TEXT,
    allowed     BOOLEAN DEFAULT TRUE,
    created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (pdf_id) REFERENCES pdfs(id),
    FOREIGN KEY (category_id) REFERENCES categories(id)
);
```

### Indexes
```sql
CREATE INDEX idx_pdfs_name ON pdfs(name);
CREATE INDEX idx_pdfs_category ON pdfs(category_id);
CREATE INDEX idx_permissions_pdf ON permissions(pdf_id);
CREATE INDEX idx_permissions_category ON permissions(category_id);
```

## Features

### 1. Search
- Command: `/search <query>`
- Searches `name`, `description`, `filename` fields
- Returns top 10 matches as embeds with buttons
- Fuzzy matching with `strings.Contains` (lowercase comparison)
- Each result shows: name, category, size, page count

### 2. Categories
- Commands: `/category create <name>`, `/category list`, `/category delete <name>`
- PDFs assigned via `/pdf category <name> <category>`
- `/list <category>` shows PDFs in that category
- Uncategorized PDFs go to "default" category

### 3. Embeds
- Every PDF response uses rich embed instead of raw file
- Shows: title (name), description, category, size, pages, upload date
- Color-coded by category
- Thumbnail: first page preview (stretch goal)

### 4. Permissions
- `/permission add <role> <pdf|category>` — grant access
- `/permission remove <role> <pdf|category>` — revoke
- `/permission list` — show all permissions
- Check logic: user role → allowed? → proceed or deny
- Default: everyone can use all commands (open by default)

### 5. Admin Commands
- `/upload` — accepts file attachment, saves to disk + DB
- `/delete <name>` — removes from disk + DB
- `/pdf info <name>` — show metadata
- `/pdf edit <name> <field> <value>` — update description/category
- Only users with Discord role "PDF Admin" can use (configurable via env var `ADMIN_ROLE`)

### 6. Pagination
- `/list` shows 10 PDFs per page
- Buttons: Previous, Page X/Y, Next
- Works with categories: `/list math` paginates math PDFs
- Search results also paginated

### 7. Metadata
- Auto-extracted: file size, page count (from PDF header)
- User-set: description, category
- Tracked: uploaded_by, created_at, updated_at
- Shown in embeds and `/pdf info`

## Data Flow

### Startup Flow
```
main.go
  → Init SQLite (migrations)
  → Scan ./pdfs/ folder
  → Sync: new files → insert DB, missing files → mark deleted
  → Load PDF cache (bytes) into memory
  → Register guild slash commands (from DB)
  → Start fsnotify watcher
  → Start Discord session
```

### Command Flow (e.g., `/search calculus`)
```
Discord interaction
  → Handler receives interaction
  → PermissionService.Check(user, command)
  → SearchService.Search("calculus")
  → SQLite query (name, description, filename LIKE %calculus%)
  → Build embeds with results
  → PaginationService.Paginate(results, page=1)
  → Respond with embeds + buttons
```

### Upload Flow (`/upload`)
```
Discord interaction (with attachment)
  → PermissionService.CheckAdmin(user)
  → Read attachment bytes
  → DiskStorage.Save(filename, bytes)
  → Extract metadata (size, pages)
  → PDFService.Create(name, path, metadata)
  → Register new slash command
  → Respond with success embed
```

### File Watcher Flow
```
fsnotify event (CREATE/MODIFY/DELETE)
  → Debounce (500ms)
  → Scan folder
  → Diff against DB
  → INSERT new / UPDATE modified / soft-DELETE removed
  → Re-register commands if changed
  → Update in-memory cache
```

## Error Handling

### Strategy
- Domain layer: return errors, never panic
- Use case layer: wrap errors with context (`fmt.Errorf("search PDFs: %w", err)`)
- Adapter layer: log errors, return user-friendly Discord messages
- Never expose internal errors to Discord users

### Domain Errors
```go
var (
    ErrPDFNotFound      = errors.New("pdf not found")
    ErrDuplicateName    = errors.New("pdf name already exists")
    ErrPermissionDenied = errors.New("permission denied")
    ErrInvalidInput     = errors.New("invalid input")
)
```

### Discord Error Responses
- Permission denied → ephemeral embed: "You don't have permission"
- Not found → ephemeral embed: "PDF not found: {name}"
- Validation → ephemeral embed: "Invalid input: {reason}"
- Internal → ephemeral embed: "Something went wrong" + log full error

### Logging
- Use Go `log/slog` (structured logging)
- Log levels: DEBUG (dev), INFO (operations), ERROR (failures)
- Log to stdout (can redirect to file)
- Include: timestamp, level, message, context (user, command, error)

### Recovery
- Bot recovers from panics in handlers (`recover()`)
- Watcher auto-restarts on error
- DB connection retry with backoff

## Command Scope

Guild commands (single server):
- Instant command sync — add PDF, command appears immediately
- No 1-hour global propagation delay
- Simpler for admin commands
- Can add global commands later if multi-server needed

## PDF Management

Hybrid model:
- Disk for bulk: drop 50 PDFs at once, bot auto-registers all
- Discord for quick: `/upload` from chat, no server access needed
- `/delete` works either way: removes from DB + disk
- Clean architecture: `StoragePort` interface abstracts both paths
