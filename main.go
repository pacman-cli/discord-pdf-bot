package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/fsnotify/fsnotify"
)

var Token = os.Getenv("DISCORD_BOT_TOKEN")
var pdfFolder = "./pdfs"

// Map command name → PDF bytes
var pdfCache = make(map[string][]byte)

// Map command name → PDF file path
var pdfFiles = make(map[string]string)
var dg *discordgo.Session

func main() {
	if Token == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable not set")
	}

	var err error
	dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
	}

	dg.AddHandler(interactionCreate)

	err = dg.Open()
	if err != nil {
		log.Fatalf("Error opening connection: %v", err)
	}

	fmt.Println("Bot is running. Press CTRL+C to exit.")

	scanPDFs()
	loadPDFsToCache()
	syncCommands()

	go watchPDFs()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}

// Sanitize PDF filenames for Discord commands
func sanitizeCommandName(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "_", "-")
	if len(name) > 32 {
		name = name[:32]
	}
	return name
}

// Scan PDF folder
func scanPDFs() {
	pdfFiles = make(map[string]string)
	files, err := ioutil.ReadDir(pdfFolder)
	if err != nil {
		log.Printf("Error reading folder: %v", err)
		return
	}
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pdf") {
			name := strings.TrimSuffix(f.Name(), ".pdf")
			cmdName := sanitizeCommandName(name)
			pdfFiles[cmdName] = filepath.Join(pdfFolder, f.Name())
		}
	}
}

// Load PDFs into memory
func loadPDFsToCache() {
	pdfCache = make(map[string][]byte)
	for cmdName, path := range pdfFiles {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Error reading PDF %s: %v", path, err)
			continue
		}
		pdfCache[cmdName] = data
	}
}

// Sync commands
func syncCommands() {
	existingCommands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		log.Printf("Error fetching existing commands: %v", err)
		return
	}

	// Delete removed commands
	for _, cmd := range existingCommands {
		if _, exists := pdfFiles[cmd.Name]; !exists {
			err := dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID)
			if err != nil {
				log.Printf("Failed to delete command '%s': %v", cmd.Name, err)
			} else {
				log.Printf("Deleted command: /%s", cmd.Name)
			}
		}
	}

	// Add new commands
	for name := range pdfFiles {
		found := false
		for _, cmd := range existingCommands {
			if cmd.Name == name {
				found = true
				break
			}
		}
		if !found {
			cmd := &discordgo.ApplicationCommand{
				Name:        name,
				Description: fmt.Sprintf("Get the %s PDF", name),
			}
			_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
			if err != nil {
				log.Printf("Cannot create command '%s': %v", name, err)
			} else {
				log.Printf("Registered command: /%s", name)
			}
		}
	}
}

// Watch PDF folder
func watchPDFs() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Watcher error: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(pdfFolder)
	if err != nil {
		log.Fatalf("Failed to watch folder: %v", err)
	}

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Write) != 0 {
				log.Printf("Folder change detected: %s", event.Name)
				scanPDFs()
				loadPDFsToCache()
				syncCommands()
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("Watcher error:", err)
		}
	}
}

// Handle slash command interactions
func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	command := i.ApplicationCommandData().Name
	data, exists := pdfCache[command]
	if !exists {
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "PDF not found!",
			},
		})
		return
	}

	// Deferred response
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Send PDF from memory
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Here is the %s PDF:", command),
		Files: []*discordgo.File{
			{
				Name:   command + ".pdf",
				Reader: bytes.NewReader(data), // <-- use bytes.NewReader
			},
		},
	})
	if err != nil {
		log.Printf("Error sending PDF: %v", err)
	}
}
