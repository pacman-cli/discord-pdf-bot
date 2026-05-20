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
