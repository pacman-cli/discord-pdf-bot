package discord

import (
	"bytes"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"discord-pdf-bot/internal/domain/entity"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func (b *Bot) handleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		b.handleCommand(s, i)
	case discordgo.InteractionMessageComponent:
		b.handleComponent(s, i)
	}
}

func (b *Bot) handleCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	command := i.ApplicationCommandData().Name

	switch command {
	case "search":
		b.handleSearch(s, i)
	case "list":
		b.handleList(s, i)
	case "upload":
		b.handleUpload(s, i)
	case "delete":
		b.handleDelete(s, i)
	case "pdf":
		subcommand := i.ApplicationCommandData().Options[0].Name
		switch subcommand {
		case "info":
			b.handlePDFInfo(s, i)
		case "edit":
			b.handlePDFEdit(s, i)
		default:
			b.respondError(s, i, "Unknown subcommand")
		}
	default:
		b.handlePDFCommand(s, i, command)
	}
}

func (b *Bot) handlePDFCommand(s *discordgo.Session, i *discordgo.InteractionCreate, name string) {
	pdf, err := b.pdfService.GetByName(name)
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("PDF not found: %s", name))
		return
	}

	if i.Member == nil {
		b.respondError(s, i, "This command can only be used in a server")
		return
	}

	userRoles := i.Member.Roles
	userID := i.Member.User.ID
	if !b.permissionService.CheckAccess(userRoles, userID, pdf.ID) {
		b.respondError(s, i, "You don't have permission to access this PDF")
		return
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	data, err := b.storageService.Read(pdf.Path)
	if err != nil {
		b.followupError(s, i, "Failed to read PDF")
		return
	}

	embed := pdfEmbed(pdf)
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files: []*discordgo.File{
			{
				Name:   pdf.Filename,
				Reader: bytes.NewReader(data),
			},
		},
	})
	if err != nil {
		slog.Error("Failed to send PDF", "error", err)
	}
}

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

func (b *Bot) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var pdfs []*entity.PDF
	var err error

	options := i.ApplicationCommandData().Options
	if len(options) > 0 && options[0].Name == "category" {
		categoryName := options[0].StringValue()
		cat, catErr := b.categoryService.GetByName(categoryName)
		if catErr != nil {
			b.respondError(s, i, fmt.Sprintf("Category not found: %s", categoryName))
			return
		}
		pdfs, err = b.pdfService.GetByCategory(cat.ID)
	} else {
		pdfs, err = b.pdfService.GetAll()
	}

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

func (b *Bot) handleUpload(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !b.isAdmin(i) {
		b.respondError(s, i, "You don't have permission to upload PDFs")
		return
	}

	attachmentID := i.ApplicationCommandData().Options[0].Value.(string)
	attachment := i.ApplicationCommandData().Resolved.Attachments[attachmentID]

	if !strings.HasSuffix(attachment.Filename, ".pdf") {
		b.respondError(s, i, "Only PDF files are allowed")
		return
	}

	// Discord attachment size limit: 25MB
	const maxUploadSize = 25 * 1024 * 1024
	if attachment.Size > maxUploadSize {
		b.respondError(s, i, "File too large. Maximum size is 25MB")
		return
	}

	description := ""
	if len(i.ApplicationCommandData().Options) > 1 {
		description = i.ApplicationCommandData().Options[1].StringValue()
	}

	data, err := s.Request("GET", attachment.URL, nil)
	if err != nil {
		b.respondError(s, i, "Failed to download attachment")
		return
	}

	name := strings.TrimSuffix(attachment.Filename, ".pdf")
	path, err := b.storageService.Save(attachment.Filename, data)
	if err != nil {
		b.respondError(s, i, "Failed to save PDF")
		return
	}

	_, err = b.pdfService.Create(name, attachment.Filename, path, description, int64(len(data)))
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("Failed to register PDF: %v", err))
		return
	}

	if err := b.SyncCommands(); err != nil {
		slog.Error("Failed to sync commands after upload", "error", err)
	}

	b.respondSuccess(s, i, fmt.Sprintf("PDF '%s' uploaded successfully", name))
}

func (b *Bot) handleDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !b.isAdmin(i) {
		b.respondError(s, i, "You don't have permission to delete PDFs")
		return
	}

	name := i.ApplicationCommandData().Options[0].StringValue()

	pdf, err := b.pdfService.GetByName(name)
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("PDF not found: %s", name))
		return
	}

	if err := b.storageService.Delete(pdf.Path); err != nil {
		slog.Error("Failed to delete PDF file", "path", pdf.Path, "error", err)
	}

	if err := b.pdfService.Delete(name); err != nil {
		b.respondError(s, i, "Failed to delete PDF from database")
		return
	}

	if err := b.SyncCommands(); err != nil {
		slog.Error("Failed to sync commands after delete", "error", err)
	}

	b.respondSuccess(s, i, fmt.Sprintf("PDF '%s' deleted", name))
}

func (b *Bot) handlePDFInfo(s *discordgo.Session, i *discordgo.InteractionCreate) {
	name := i.ApplicationCommandData().Options[0].StringValue()

	pdf, err := b.pdfService.GetByName(name)
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("PDF not found: %s", name))
		return
	}

	embed := pdfEmbed(pdf)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (b *Bot) handlePDFEdit(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !b.isAdmin(i) {
		b.respondError(s, i, "You don't have permission to edit PDFs")
		return
	}

	name := i.ApplicationCommandData().Options[0].StringValue()
	field := i.ApplicationCommandData().Options[1].StringValue()
	value := i.ApplicationCommandData().Options[2].StringValue()

	pdf, err := b.pdfService.GetByName(name)
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("PDF not found: %s", name))
		return
	}

	switch field {
	case "description":
		pdf.Description = value
	case "category":
		cat, err := b.categoryService.GetByName(value)
		if err != nil {
			b.respondError(s, i, fmt.Sprintf("Category not found: %s", value))
			return
		}
		pdf.CategoryID = &cat.ID
	default:
		b.respondError(s, i, fmt.Sprintf("Unknown field: %s", field))
		return
	}

	if err := b.pdfService.Update(pdf); err != nil {
		b.respondError(s, i, "Failed to update PDF")
		return
	}

	b.respondSuccess(s, i, fmt.Sprintf("PDF '%s' updated", pdf.Name))
}

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

	state.Timestamp = time.Now()

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

func (b *Bot) isAdmin(i *discordgo.InteractionCreate) bool {
	for _, roleID := range i.Member.Roles {
		role, err := b.session.State.Role(i.GuildID, roleID)
		if err != nil {
			continue
		}
		if role.Name == b.adminRole {
			return true
		}
	}
	return false
}

func (b *Bot) respondError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := errorEmbed(message)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) respondSuccess(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := successEmbed(message)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func (b *Bot) followupError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := errorEmbed(message)
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
}
