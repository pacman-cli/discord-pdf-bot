package discord

import (
	"fmt"
	"log/slog"

	"discord-pdf-bot/internal/usecase"

	"github.com/bwmarrin/discordgo"
)

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

func NewBot(
	token string,
	pdfService *usecase.PDFService,
	categoryService *usecase.CategoryService,
	permissionService *usecase.PermissionService,
	storageService *usecase.StorageService,
	guildID string,
	adminRole string,
) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("create discord session: %w", err)
	}

	bot := &Bot{
		session:           dg,
		pdfService:        pdfService,
		categoryService:   categoryService,
		permissionService: permissionService,
		storageService:    storageService,
		guildID:           guildID,
		adminRole:         adminRole,
	}

	bot.pagination = newPaginationCache()

	dg.AddHandler(bot.handleInteraction)

	return bot, nil
}

func (b *Bot) Open() error {
	return b.session.Open()
}

func (b *Bot) Close() error {
	return b.session.Close()
}

func (b *Bot) SyncCommands() error {
	existing, err := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	if err != nil {
		return fmt.Errorf("fetch commands: %w", err)
	}

	pdfs, err := b.pdfService.GetAll()
	if err != nil {
		return fmt.Errorf("fetch pdfs: %w", err)
	}

	existingMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existing {
		existingMap[cmd.Name] = cmd
	}

	// Delete commands for removed PDFs
	for _, cmd := range existing {
		found := false
		for _, pdf := range pdfs {
			if pdf.Name == cmd.Name {
				found = true
				break
			}
		}
		if !found {
			err := b.session.ApplicationCommandDelete(b.session.State.User.ID, b.guildID, cmd.ID)
			if err != nil {
				slog.Error("Failed to delete command", "command", cmd.Name, "error", err)
			} else {
				slog.Info("Deleted command", "command", cmd.Name)
			}
		}
	}

	// Add commands for new PDFs
	for _, pdf := range pdfs {
		if _, exists := existingMap[pdf.Name]; !exists {
			cmd := &discordgo.ApplicationCommand{
				Name:        pdf.Name,
				Description: fmt.Sprintf("Get the %s PDF", pdf.Name),
			}
			_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.guildID, cmd)
			if err != nil {
				slog.Error("Failed to create command", "command", pdf.Name, "error", err)
			} else {
				slog.Info("Registered command", "command", pdf.Name)
			}
		}
	}

	b.registerUtilityCommands()

	return nil
}

func (b *Bot) registerUtilityCommands() {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "search",
			Description: "Search for PDFs",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "query",
					Description: "Search query",
					Required:    true,
				},
			},
		},
		{
			Name:        "list",
			Description: "List all PDFs",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "category",
					Description: "Filter by category",
					Required:    false,
				},
			},
		},
		{
			Name:        "upload",
			Description: "Upload a PDF file",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionAttachment,
					Name:        "file",
					Description: "PDF file to upload",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "PDF description",
					Required:    false,
				},
			},
		},
		{
			Name:        "delete",
			Description: "Delete a PDF",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "name",
					Description: "PDF name to delete",
					Required:    true,
				},
			},
		},
		{
			Name:        "pdf",
			Description: "PDF management",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "info",
					Description: "Show PDF info",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "PDF name",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "edit",
					Description: "Edit PDF metadata",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "PDF name",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "field",
							Description: "Field to edit (description, category)",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "value",
							Description: "New value",
							Required:    true,
						},
					},
				},
			},
		},
	}

	existing, _ := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	existingNames := make(map[string]bool)
	for _, e := range existing {
		existingNames[e.Name] = true
	}

	for _, cmd := range commands {
		if !existingNames[cmd.Name] {
			_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.guildID, cmd)
			if err != nil {
				slog.Error("Failed to create command", "command", cmd.Name, "error", err)
			}
		}
	}
}
