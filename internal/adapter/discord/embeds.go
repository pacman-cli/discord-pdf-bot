package discord

import (
	"fmt"

	"discord-pdf-bot/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
)

func pdfEmbed(pdf *entity.PDF) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       pdf.Name,
		Description: pdf.Description,
		Color:       0x5865F2,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "File", Value: pdf.Filename, Inline: true},
			{Name: "Size", Value: formatFileSize(pdf.FileSize), Inline: true},
			{Name: "Pages", Value: fmt.Sprintf("%d", pdf.PageCount), Inline: true},
		},
	}

	return embed
}

func errorEmbed(message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: message,
		Color:       0xED4245,
	}
}

func successEmbed(message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Success",
		Description: message,
		Color:       0x57F287,
	}
}

func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
