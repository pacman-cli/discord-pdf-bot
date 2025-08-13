// package main
//
// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"path/filepath"
// 	"strings"
// 	"syscall"
//
// 	"github.com/bwmarrin/discordgo"
// 	"github.com/fsnotify/fsnotify"
// )
//
// // <-- Your bot token here
// var Token = "REDACTED"
//
// // Folder where PDFs are stored
// var pdfFolder = "./pdfs"
//
// // Map command name → PDF file path
// var pdfFiles = make(map[string]string)
//
// // Discord session
// var dg *discordgo.Session
//
// func main() {
// 	var err error
// 	dg, err = discordgo.New("Bot " + Token)
// 	if err != nil {
// 		log.Fatalf("Error creating Discord session: %v", err)
// 	}
//
// 	dg.AddHandler(interactionCreate)
//
// 	err = dg.Open()
// 	if err != nil {
// 		log.Fatalf("Error opening connection: %v", err)
// 	}
//
// 	fmt.Println("Bot is running. Press CTRL+C to exit.")
//
// 	// Initial scan and register commands
// 	scanPDFs()
// 	syncCommands()
//
// 	// Start watching folder for changes
// 	go watchPDFs()
//
// 	// Wait for termination signal
// 	stop := make(chan os.Signal, 1)
// 	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
// 	<-stop
// }
//
// // Scan the PDF folder and populate pdfFiles map
// func scanPDFs() {
// 	pdfFiles = make(map[string]string)
// 	err := filepath.Walk(pdfFolder, func(path string, info os.FileInfo, err error) error {
// 		if err != nil {
// 			return err
// 		}
// 		if !info.IsDir() && strings.HasSuffix(info.Name(), ".pdf") {
// 			cmdName := strings.TrimSuffix(info.Name(), ".pdf")
// 			pdfFiles[cmdName] = path
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		log.Printf("Error scanning PDFs: %v", err)
// 	}
// }
//
// // Synchronize slash commands with PDF files
// func syncCommands() {
// 	existingCommands, err := dg.ApplicationCommands(dg.State.User.ID, "")
// 	if err != nil {
// 		log.Printf("Error fetching existing commands: %v", err)
// 		return
// 	}
//
// 	// Delete commands for PDFs that no longer exist
// 	for _, cmd := range existingCommands {
// 		if _, exists := pdfFiles[cmd.Name]; !exists {
// 			err := dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID)
// 			if err != nil {
// 				log.Printf("Failed to delete command '%s': %v", cmd.Name, err)
// 			} else {
// 				log.Printf("Deleted command: /%s", cmd.Name)
// 			}
// 		}
// 	}
//
// 	// Create commands for new PDFs
// 	for cmdName := range pdfFiles {
// 		found := false
// 		for _, cmd := range existingCommands {
// 			if cmd.Name == cmdName {
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			cmd := &discordgo.ApplicationCommand{
// 				Name:        cmdName,
// 				Description: fmt.Sprintf("Get the %s PDF", cmdName),
// 			}
// 			_, err := dg.ApplicationCommandCreate(dg.State.User.ID, "", cmd)
// 			if err != nil {
// 				log.Printf("Cannot create command '%s': %v", cmdName, err)
// 			} else {
// 				log.Printf("Registered command: /%s", cmdName)
// 			}
// 		}
// 	}
// }
//
// // Watch the pdfs folder for changes
// func watchPDFs() {
// 	watcher, err := fsnotify.NewWatcher()
// 	if err != nil {
// 		log.Fatalf("Failed to create watcher: %v", err)
// 	}
// 	defer watcher.Close()
//
// 	err = watcher.Add(pdfFolder)
// 	if err != nil {
// 		log.Fatalf("Failed to watch folder: %v", err)
// 	}
//
// 	for {
// 		select {
// 		case event, ok := <-watcher.Events:
// 			if !ok {
// 				return
// 			}
// 			// Only care about create or remove events
// 			if event.Op&(fsnotify.Create|fsnotify.Remove) != 0 {
// 				log.Printf("Folder change detected: %s", event.Name)
// 				scanPDFs()
// 				syncCommands()
// 			}
// 		case err, ok := <-watcher.Errors:
// 			if !ok {
// 				return
// 			}
// 			log.Println("Watcher error:", err)
// 		}
// 	}
// }
//
// // Handle slash command interactions
//
// func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
// 	if i.Type != discordgo.InteractionApplicationCommand {
// 		return
// 	}
//
// 	command := i.ApplicationCommandData().Name
// 	filePath, exists := pdfFiles[command]
// 	if !exists {
// 		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 			Type: discordgo.InteractionResponseChannelMessageWithSource,
// 			Data: &discordgo.InteractionResponseData{
// 				Content: "PDF not found!",
// 			},
// 		})
// 		return
// 	}
//
// 	// Send a deferred response to give more time
// 	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
// 		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
// 	})
//
// 	// Open the PDF
// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		log.Printf("Error opening PDF '%s': %v", filePath, err)
// 		s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
// 			Content: fmt.Sprintf("Error: could not open '%s'", filePath),
// 		})
// 		return
// 	}
// 	defer file.Close()
//
// 	// Send the PDF as a follow-up
// 	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
// 		Content: fmt.Sprintf("Here is the %s PDF:", command),
// 		Files: []*discordgo.File{
// 			{
// 				Name:   filepath.Base(filePath),
// 				Reader: file,
// 			},
// 		},
// 	})
// 	if err != nil {
// 		log.Printf("Error sending PDF: %v", err)
// 	}
// }

package main

import (
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

// <-- Your bot token here
// var Token = "REDACTED"

var Token = os.Getenv("DISCORD_BOT_TOKEN")
var pdfFolder = "./pdfs"

// Map command name → PDF bytes
var pdfCache = make(map[string][]byte)

// Map command name → PDF file path (for logging or reload)
var pdfFiles = make(map[string]string)

var dg *discordgo.Session

func main() {
	Token = os.Getenv("DISCORD_BOT_TOKEN")
	if Token == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable not set")
	}

	var err error
	dg, err = discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %v", err)
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

	// Initial scan and load PDFs into memory
	scanPDFs()
	loadPDFsToCache()
	syncCommands()

	// Watch pdfs folder for changes
	go watchPDFs()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}

// Scan PDF folder for files
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
			pdfFiles[name] = filepath.Join(pdfFolder, f.Name())
		}
	}
}

// Load PDFs into memory
func loadPDFsToCache() {
	pdfCache = make(map[string][]byte)
	for name, path := range pdfFiles {
		data, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Error reading PDF %s: %v", path, err)
			continue
		}
		pdfCache[name] = data
	}
}

// Sync slash commands with current PDF files
func syncCommands() {
	existingCommands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		log.Printf("Error fetching existing commands: %v", err)
		return
	}

	// Delete commands for removed PDFs
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

	// Add commands for new PDFs
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

// Watch pdfs folder for changes
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

	// Send deferred response first
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Send the PDF from memory
	_, err := s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Content: fmt.Sprintf("Here is the %s PDF:", command),
		Files: []*discordgo.File{
			{
				Name:   command + ".pdf",
				Reader: strings.NewReader(string(data)),
			},
		},
	})
	if err != nil {
		log.Printf("Error sending PDF: %v", err)
	}
}
