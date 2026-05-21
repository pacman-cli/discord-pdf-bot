# Pagination Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add button-based pagination to `/list` and `/search` commands with 10 items per page and in-memory state tracking.

**Architecture:** New `pagination.go` file holds state struct, cache, and embed/button builders. Handlers in `handlers.go` create pagination state on initial response, button handler edits the embed on page change. UUID key per interaction, 15-min TTL with lazy cleanup.

**Tech Stack:** Go, discordgo, google/uuid (already in go.mod as indirect dep)

---

## File Structure

| File | Action | Purpose |
|------|--------|---------|
| `internal/adapter/discord/pagination.go` | Create | PaginationState, cache, embed/button builders, button handler |
| `internal/adapter/discord/pagination_test.go` | Create | Unit tests for pagination helpers |
| `internal/adapter/discord/handlers.go` | Modify | Wire pagination into handleList, handleSearch, handleComponent |
| `internal/adapter/discord/embeds.go` | Modify | Remove searchResultEmbed (replaced by buildPageEmbed) |

---

### Task 1: Create pagination.go

**Files:**
- Create: `internal/adapter/discord/pagination.go`

- [ ] **Step 1: Create pagination.go with state, cache, and helpers**

```go
package discord

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"discord-pdf-bot/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

const (
	pageSize     = 10
	cacheTTL     = 15 * time.Minute
	customPrefix = "page:"
)

type PaginationState struct {
	AllPDFs    []*entity.PDF
	Query      string // "" for list, search term for search
	Page       int    // 0-indexed
	TotalPages int
	Timestamp  time.Time
}

type paginationCache struct {
	mu    sync.RWMutex
	items map[string]*PaginationState
}

func newPaginationCache() *paginationCache {
	return &paginationCache{items: make(map[string]*PaginationState)}
}

func (c *paginationCache) set(key string, state *PaginationState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = state
}

func (c *paginationCache) get(key string) (*PaginationState, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	state, ok := c.items[key]
	if !ok {
		return nil, false
	}
	if time.Since(state.Timestamp) > cacheTTL {
		delete(c.items, key)
		return nil, false
	}
	return state, true
}

func (c *paginationCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if time.Since(v.Timestamp) > cacheTTL {
			delete(c.items, k)
		}
	}
}

func totalPages(count int) int {
	if count == 0 {
		return 0
	}
	return (count + pageSize - 1) / pageSize
}

func buildPageEmbed(title string, pdfs []*entity.PDF, page, total int) *discordgo.MessageEmbed {
	start := page * pageSize
	end := start + pageSize
	if end > len(pdfs) {
		end = len(pdfs)
	}

	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("Page %d of %d (%d total)", page+1, total, len(pdfs)),
		Color:       0x5865F2,
	}

	for i := start; i < end; i++ {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d. %s", i+1, pdfs[i].Name),
			Value:  pdfs[i].Description,
			Inline: false,
		})
	}

	return embed
}

func buildPaginationButtons(stateKey string, page, total int) []discordgo.MessageComponent {
	prevDisabled := page <= 0
	nextDisabled := page >= total-1

	return []discordgo.MessageComponent{
		discordgo.ActionsRow{
			Components: []discordgo.MessageComponent{
				discordgo.Button{
					Label:    "Previous",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("%s%s:prev", customPrefix, stateKey),
					Disabled: prevDisabled,
				},
				discordgo.Button{
					Label:    "Next",
					Style:    discordgo.PrimaryButton,
					CustomID: fmt.Sprintf("%s%s:next", customPrefix, stateKey),
					Disabled: nextDisabled,
				},
			},
		},
	}
}

func parsePageCustomID(customID string) (stateKey string, direction string, ok bool) {
	if !strings.HasPrefix(customID, customPrefix) {
		return "", "", false
	}
	rest := strings.TrimPrefix(customID, customPrefix)
	parts := strings.Split(rest, ":")
	if len(parts) != 2 {
		return "", "", false
	}
	return parts[0], parts[1], true
}
```

- [ ] **Step 2: Build to verify syntax**

Run: `go build ./internal/adapter/discord/`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/discord/pagination.go
git commit -m "feat: add pagination state and helpers"
```

---

### Task 2: Add pagination tests

**Files:**
- Create: `internal/adapter/discord/pagination_test.go`

- [ ] **Step 1: Write tests for totalPages, buildPageEmbed, parsePageCustomID**

```go
package discord

import (
	"fmt"
	"testing"
	"time"

	"discord-pdf-bot/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
)

func makePDFs(n int) []*entity.PDF {
	pdfs := make([]*entity.PDF, n)
	for i := range pdfs {
		pdfs[i] = &entity.PDF{
			Name:        fmt.Sprintf("pdf_%d", i+1),
			Description: fmt.Sprintf("Description %d", i+1),
		}
	}
	return pdfs
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		count    int
		expected int
	}{
		{0, 0},
		{1, 1},
		{10, 1},
		{11, 2},
		{25, 3},
	}
	for _, tt := range tests {
		got := totalPages(tt.count)
		if got != tt.expected {
			t.Errorf("totalPages(%d) = %d, want %d", tt.count, got, tt.expected)
		}
	}
}

func TestBuildPageEmbed(t *testing.T) {
	pdfs := makePDFs(15)

	// Page 0: items 1-10
	embed := buildPageEmbed("Test", pdfs, 0, 2)
	if len(embed.Fields) != 10 {
		t.Errorf("page 0: got %d fields, want 10", len(embed.Fields))
	}

	// Page 1: items 11-15
	embed = buildPageEmbed("Test", pdfs, 1, 2)
	if len(embed.Fields) != 5 {
		t.Errorf("page 1: got %d fields, want 5", len(embed.Fields))
	}
}

func TestBuildPageEmbedSinglePage(t *testing.T) {
	pdfs := makePDFs(5)
	embed := buildPageEmbed("Test", pdfs, 0, 1)
	if len(embed.Fields) != 5 {
		t.Errorf("single page: got %d fields, want 5", len(embed.Fields))
	}
}

func TestBuildPaginationButtons(t *testing.T) {
	// First page: prev disabled
	buttons := buildPaginationButtons("abc", 0, 3)
	row := buttons[0].(discordgo.ActionsRow)
	prev := row.Components[0].(discordgo.Button)
	next := row.Components[1].(discordgo.Button)
	if !prev.Disabled {
		t.Error("prev should be disabled on first page")
	}
	if next.Disabled {
		t.Error("next should not be disabled on first page")
	}

	// Last page: next disabled
	buttons = buildPaginationButtons("abc", 2, 3)
	row = buttons[0].(discordgo.ActionsRow)
	prev = row.Components[0].(discordgo.Button)
	next = row.Components[1].(discordgo.Button)
	if prev.Disabled {
		t.Error("prev should not be disabled on last page")
	}
	if !next.Disabled {
		t.Error("next should be disabled on last page")
	}

	// Single page: both disabled
	buttons = buildPaginationButtons("abc", 0, 1)
	row = buttons[0].(discordgo.ActionsRow)
	prev = row.Components[0].(discordgo.Button)
	next = row.Components[1].(discordgo.Button)
	if !prev.Disabled {
		t.Error("prev should be disabled on single page")
	}
	if !next.Disabled {
		t.Error("next should be disabled on single page")
	}
}

func TestParsePageCustomID(t *testing.T) {
	key, dir, ok := parsePageCustomID("page:abc123:next")
	if !ok || key != "abc123" || dir != "next" {
		t.Errorf("got key=%s dir=%s ok=%v, want abc123/next/true", key, dir, ok)
	}

	_, _, ok = parsePageCustomID("other:abc:prev")
	if ok {
		t.Error("should fail for non-page prefix")
	}

	_, _, ok = parsePageCustomID("page:abc")
	if ok {
		t.Error("should fail for malformed id")
	}
}

func TestPaginationCache(t *testing.T) {
	c := newPaginationCache()

	c.set("k1", &PaginationState{Page: 0})
	state, ok := c.get("k1")
	if !ok || state.Page != 0 {
		t.Error("expected to get k1")
	}

	_, ok = c.get("missing")
	if ok {
		t.Error("should not find missing key")
	}

	// Expired entry
	c.set("old", &PaginationState{Timestamp: time.Now().Add(-cacheTTL - time.Minute)})
	_, ok = c.get("old")
	if ok {
		t.Error("expired entry should not be found")
	}
}
```

- [ ] **Step 2: Run tests**

Run: `go test ./internal/adapter/discord/ -v -run "TestTotalPages|TestBuildPage|TestBuildPagination|TestParsePage|TestPaginationCache"`
Expected: all PASS

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/discord/pagination_test.go
git commit -m "test: add pagination helper tests"
```

---

### Task 3: Wire pagination into handleList and handleSearch

**Files:**
- Modify: `internal/adapter/discord/handlers.go:86-142`
- Modify: `internal/adapter/discord/bot.go:12-20`

- [ ] **Step 1: Add paginationCache to Bot struct**

In `bot.go`, add `pagination *paginationCache` to the Bot struct and initialize it in `NewBot`:

```go
type Bot struct {
	session           *discordgo.Session
	pdfService        *usecase.PDFService
	categoryService   *usecase.CategoryService
	permissionService *usecase.PermissionService
	storageService    *usecase.StorageService
	pagination        *paginationCache
	guildID           string
	adminRole         string
}
```

In `NewBot`, after creating the bot struct, add:

```go
bot.pagination = newPaginationCache()
```

- [ ] **Step 2: Replace handleSearch with paginated version**

Replace the entire `handleSearch` function in `handlers.go` (lines 86-107):

```go
func (b *Bot) handleSearch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	query := i.ApplicationCommandData().Options[0].StringValue()

	results, err := b.pdfService.Search(query)
	if err != nil {
		b.respondError(s, i, "Search failed")
		return
	}

	if len(results) == 0 {
		b.respondError(s, i, "No PDFs found")
		return
	}

	total := totalPages(len(results))
	stateKey := uuid.New().String()
	b.pagination.set(stateKey, &PaginationState{
		AllPDFs:    results,
		Query:      query,
		Page:       0,
		TotalPages: total,
		Timestamp:  time.Now(),
	})

	embed := buildPageEmbed("Search Results", results, 0, total)
	components := buildPaginationButtons(stateKey, 0, total)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}
```

- [ ] **Step 3: Replace handleList with paginated version**

Replace the entire `handleList` function in `handlers.go` (lines 109-142):

```go
func (b *Bot) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	pdfs, err := b.pdfService.GetAll()
	if err != nil {
		b.respondError(s, i, "Failed to fetch PDFs")
		return
	}

	if len(pdfs) == 0 {
		b.respondError(s, i, "No PDFs available")
		return
	}

	total := totalPages(len(pdfs))
	stateKey := uuid.New().String()
	b.pagination.set(stateKey, &PaginationState{
		AllPDFs:    pdfs,
		Page:       0,
		TotalPages: total,
		Timestamp:  time.Now(),
	})

	embed := buildPageEmbed("Available PDFs", pdfs, 0, total)
	components := buildPaginationButtons(stateKey, 0, total)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}
```

- [ ] **Step 4: Add missing imports to handlers.go**

Add to the import block in `handlers.go`:

```go
"time"

"github.com/google/uuid"
```

- [ ] **Step 5: Build to verify**

Run: `go build ./...`
Expected: no output (success)

- [ ] **Step 6: Commit**

```bash
git add internal/adapter/discord/bot.go internal/adapter/discord/handlers.go
git commit -m "feat: wire pagination into /list and /search handlers"
```

---

### Task 4: Implement button handler

**Files:**
- Modify: `internal/adapter/discord/handlers.go:284-293`

- [ ] **Step 1: Replace handleComponent with pagination routing**

Replace the `handleComponent` function in `handlers.go` (lines 284-293):

```go
func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID

	stateKey, direction, ok := parsePageCustomID(customID)
	if !ok {
		b.respondError(s, i, "Unknown component")
		return
	}

	state, ok := b.pagination.get(stateKey)
	if !ok {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "Pagination expired. Run the command again.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch direction {
	case "next":
		if state.Page < state.TotalPages-1 {
			state.Page++
		}
	case "prev":
		if state.Page > 0 {
			state.Page--
		}
	}

	var title string
	if state.Query != "" {
		title = "Search Results"
	} else {
		title = "Available PDFs"
	}

	embed := buildPageEmbed(title, state.AllPDFs, state.Page, state.TotalPages)
	components := buildPaginationButtons(stateKey, state.Page, state.TotalPages)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		},
	})
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/discord/handlers.go
git commit -m "feat: implement pagination button handler"
```

---

### Task 5: Remove old searchResultEmbed

**Files:**
- Modify: `internal/adapter/discord/embeds.go:32-49`

- [ ] **Step 1: Delete searchResultEmbed function**

Remove the `searchResultEmbed` function from `embeds.go` (lines 32-49). It's replaced by `buildPageEmbed` in `pagination.go`.

- [ ] **Step 2: Build to verify no references remain**

Run: `go build ./...`
Expected: no output (success)

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/discord/embeds.go
git commit -m "refactor: remove unused searchResultEmbed"
```

---

### Task 6: Run all tests

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: all tests pass, including new pagination tests

- [ ] **Step 2: Final commit if needed**
