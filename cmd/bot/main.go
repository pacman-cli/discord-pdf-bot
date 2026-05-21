package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"discord-pdf-bot/internal/adapter/discord"
	"discord-pdf-bot/internal/adapter/repository"
	"discord-pdf-bot/internal/adapter/storage"
	"discord-pdf-bot/internal/infrastructure/database"
	"discord-pdf-bot/internal/infrastructure/watcher"
	"discord-pdf-bot/internal/usecase"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		slog.Error("DISCORD_BOT_TOKEN environment variable not set")
		os.Exit(1)
	}

	guildID := os.Getenv("GUILD_ID")
	if guildID == "" {
		slog.Error("GUILD_ID environment variable not set")
		os.Exit(1)
	}

	adminRole := os.Getenv("ADMIN_ROLE")
	if adminRole == "" {
		adminRole = "PDF Admin"
	}

	// Database
	db, err := database.NewSQLite("./data/bot.db")
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		slog.Error("Failed to migrate database", "error", err)
		os.Exit(1)
	}

	// Repositories
	pdfRepo := repository.NewSQLitePDFRepository(db.DB())
	categoryRepo := repository.NewSQLiteCategoryRepository(db.DB())
	permissionRepo := repository.NewSQLitePermissionRepository(db.DB())

	// Storage
	diskStorage := storage.NewDiskStorage("./pdfs")

	// Services
	pdfService := usecase.NewPDFService(pdfRepo)
	categoryService := usecase.NewCategoryService(categoryRepo)
	permissionService := usecase.NewPermissionService(permissionRepo)

	// Initial sync from disk
	files, err := diskStorage.List("./pdfs")
	if err != nil {
		slog.Error("Failed to list PDFs", "error", err)
		os.Exit(1)
	}

	if err := pdfService.SyncFromDisk(files); err != nil {
		slog.Error("Failed to sync PDFs from disk", "error", err)
		os.Exit(1)
	}

	// Bot
	bot, err := discord.NewBot(token, pdfService, categoryService, permissionService, diskStorage, guildID, adminRole)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	if err := bot.Open(); err != nil {
		slog.Error("Failed to open bot connection", "error", err)
		os.Exit(1)
	}
	defer bot.Close()

	if err := bot.SyncCommands(); err != nil {
		slog.Error("Failed to sync commands", "error", err)
		os.Exit(1)
	}

	// File watcher
	onChange := func() {
		slog.Info("PDF folder changed, syncing...")
		files, err := diskStorage.List("./pdfs")
		if err != nil {
			slog.Error("Failed to list PDFs", "error", err)
			return
		}

		if err := pdfService.SyncFromDisk(files); err != nil {
			slog.Error("Failed to sync PDFs", "error", err)
			return
		}

		if err := bot.SyncCommands(); err != nil {
			slog.Error("Failed to sync commands", "error", err)
		}
	}

	fw, err := watcher.NewFileWatcher("./pdfs", onChange)
	if err != nil {
		slog.Error("Failed to create file watcher", "error", err)
		os.Exit(1)
	}
	defer fw.Close()

	slog.Info("Bot is running", "guild", guildID)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down...")
}
