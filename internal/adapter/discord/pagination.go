package discord

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"discord-pdf-bot/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
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
	c := &paginationCache{items: make(map[string]*PaginationState)}
	go c.cleanupLoop()
	return c
}

func (c *paginationCache) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.cleanup()
	}
}

func (c *paginationCache) set(key string, state *PaginationState) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if state.Timestamp.IsZero() {
		state.Timestamp = time.Now()
	}
	c.items[key] = state
}

func (c *paginationCache) get(key string) (*PaginationState, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
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
