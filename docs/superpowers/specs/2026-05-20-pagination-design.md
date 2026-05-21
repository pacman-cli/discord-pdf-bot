# Pagination Design

**Date:** 2026-05-20
**Status:** Approved

## Problem

The `/list` and `/search` commands return all results in a single embed. Discord limits embeds to 25 fields, so large PDF collections get truncated. Users need a way to browse results in pages.

## Solution

Button-based pagination with Previous/Next buttons below the embed. 10 items per page. In-memory state tracking per message.

## Design

### State Management

```go
type PaginationState struct {
    Query      string    // "" for list, search term for search
    Page       int       // current page (0-indexed)
    TotalPages int
    Timestamp  time.Time // for TTL cleanup
}
```

State stored in `map[string]*PaginationState` on Bot struct, keyed by UUID generated per interaction. Protected by `sync.RWMutex`. TTL: 15 minutes with lazy cleanup on button press.

### Button Custom IDs

Format: `page:{stateKey}:{direction}` where stateKey is the UUID and direction is `prev` or `next`.

### Components

**New file: `internal/adapter/discord/pagination.go`**
- `PaginationState` struct
- `buildPageEmbed(pdfs []*entity.PDF, page, total int) *discordgo.MessageEmbed` — shows page N of M, 10 items with numbers
- `buildPaginationButtons(stateKey string, page, total int) []discordgo.MessageComponent` — Previous/Next buttons, disabled at boundaries
- `handlePagination(s, i)` — reads custom_id, looks up state, updates page, edits embed

**Modified: `internal/adapter/discord/handlers.go`**
- `handleList` — fetch all PDFs, create pagination state, send embed with buttons
- `handleSearch` — fetch search results, create pagination state, send embed with buttons
- `handleComponent` — route `page:` prefix to `handlePagination`

**Modified: `internal/adapter/discord/embeds.go`**
- `searchResultEmbed` replaced by `buildPageEmbed`

### Data Flow

1. User runs `/list` or `/search query:math`
2. Bot fetches all results, calculates total pages
3. Bot generates UUID, stores PaginationState in cache keyed by UUID
4. Bot sends embed with page 1 + Previous(disabled)/Next buttons
5. User clicks Next → bot reads custom_id, looks up state, increments page
6. Bot edits original embed with new page, updates button states
7. At boundaries, corresponding button is disabled

### Edge Cases

- **Empty results**: no buttons, same as current
- **Single page**: both buttons disabled
- **Stale buttons** (bot restarted): respond ephemeral "Pagination expired. Run the command again."
- **Concurrent clicks**: mutex protects cache; Discord serializes button interactions per message

### Files

| File | Action |
|------|--------|
| `internal/adapter/discord/pagination.go` | Create |
| `internal/adapter/discord/handlers.go` | Modify |
| `internal/adapter/discord/embeds.go` | Modify |
