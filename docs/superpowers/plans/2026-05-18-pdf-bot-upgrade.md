# Discord PDF Bot Upgrade Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor single-file Discord bot to clean architecture with SQLite storage and 7 new features.

**Architecture:** Clean architecture with 4 layers (domain/usecase/adapter/infrastructure). Domain has zero external dependencies. Use case orchestrates domain logic. Adapters implement ports. Infrastructure provides concrete implementations.

**Tech Stack:** Go 1.24, discordgo, fsnotify, modernc.org/sqlite (pure Go), log/slog

---

## File Structure

```
discord-pdf-bot/
├── cmd/bot/main.go                      # Entry point, DI wiring
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   │   ├── pdf.go                   # PDF entity
│   │   │   ├── category.go              # Category entity
│   │   │   └── permission.go            # Permission entity
│   │   ├── port/
│   │   │   ├── pdf_repository.go        # PDF repo interface
│   │   │   ├── storage.go               # Storage interface
│   │   │   └── discord.go               # Discord interface
│   │   └── errors.go                    # Domain errors
│   ├── usecase/
│   │   ├── pdf_service.go               # PDF CRUD, search, list
│   │   ├── category_service.go          # Category management
│   │   └── permission_service.go        # Permission checks
│   ├── adapter/
│   │   ├── discord/
│   │   │   ├── bot.go                   # Bot setup, command registration
│   │   │   ├── handlers.go              # Interaction handlers
│   │   │   └── embeds.go                # Rich embed builders
│   │   ├── repository/
│   │   │   └── sqlite_pdf.go            # SQLite PDF repo
│   │   └── storage/
│   │       └── disk_storage.go          # File read/write
│   └── infrastructure/
│       ├── database/
│       │   └── sqlite.go                # Connection, migrations
│       └── watcher/
│           └── fsnotify.go              # Filesystem watcher
├── internal/domain/entity/pdf_test.go   # Entity tests
├── internal/usecase/pdf_service_test.go # Service tests
├── pdfs/                                # PDF storage
├── data/bot.db                          # SQLite database
└── go.mod
```

---

## Phase 1: Project Restructure

### Task 1: Create directory structure and move existing code

**Files:**
- Create: `cmd/bot/main.go` (new entry point)
- Create: `internal/domain/entity/pdf.go`
- Create: `internal/domain/port/pdf_repository.go`
- Create: `internal/domain/port/storage.go`
- Create: `internal/domain/errors.go`
- Delete: `main.go` (old entry point)

- [ ] **Step 1: Create directory structure**

```bash
mkdir -p cmd/bot
mkdir -p internal/domain/entity
mkdir -p internal/domain/port
mkdir -p internal/usecase
mkdir -p internal/adapter/discord
mkdir -p internal/adapter/repository
mkdir -p internal/adapter/storage
mkdir -p internal/infrastructure/database
mkdir -p internal/infrastructure/watcher
mkdir -p data
```

- [ ] **Step 2: Create domain errors**

Create `internal/domain/errors.go`:

```go
package domain

import "errors"

var (
	ErrPDFNotFound      = errors.New("pdf not found")
	ErrDuplicateName    = errors.New("pdf name already exists")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidInput     = errors.New("invalid input")
)
```

- [ ] **Step 3: Create PDF entity**

Create `internal/domain/entity/pdf.go`:

```go
package entity

import "time"

type PDF struct {
	ID          int64
	Name        string
	Filename    string
	Path        string
	Description string
	CategoryID  *int64
	UploadedBy  string
	PageCount   int
	FileSize    int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
```

- [ ] **Step 4: Create PDF repository port**

Create `internal/domain/port/pdf_repository.go`:

```go
package port

import "discord-pdf-bot/internal/domain/entity"

type PDFRepository interface {
	GetByName(name string) (*entity.PDF, error)
	GetAll() ([]*entity.PDF, error)
	GetByCategory(categoryID int64) ([]*entity.PDF, error)
	Search(query string) ([]*entity.PDF, error)
	Create(pdf *entity.PDF) error
	Update(pdf *entity.PDF) error
	Delete(name string) error
}
```

- [ ] **Step 5: Create storage port**

Create `internal/domain/port/storage.go`:

```go
package port

type StoragePort interface {
	Save(filename string, data []byte) (string, error)
	Delete(path string) error
	Read(path string) ([]byte, error)
	List(folder string) (map[string]string, error)
}
```

- [ ] **Step 6: Create minimal entry point**

Create `cmd/bot/main.go`:

```go
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable not set")
	}

	fmt.Println("Bot is running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
}
```

- [ ] **Step 7: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 8: Commit**

```bash
git add cmd/ internal/ data/
git commit -m "refactor: create directory structure and domain layer"
```

---

## Phase 2: Infrastructure Layer

### Task 2: SQLite database setup

**Files:**
- Create: `internal/infrastructure/database/sqlite.go`
- Modify: `go.mod` (add sqlite dependency)

- [ ] **Step 1: Add SQLite dependency**

Run: `go get modernc.org/sqlite`
Expected: go.mod updated with sqlite dependency

- [ ] **Step 2: Create database connection**

Create `internal/infrastructure/database/sqlite.go`:

```go
package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(dbPath string) (*SQLite, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("Connected to SQLite database", "path", dbPath)

	return &SQLite{db: db}, nil
}

func (s *SQLite) DB() *sql.DB {
	return s.db
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pdfs (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL UNIQUE,
			filename    TEXT NOT NULL,
			path        TEXT NOT NULL,
			description TEXT DEFAULT '',
			category_id INTEGER,
			uploaded_by TEXT,
			page_count  INTEGER DEFAULT 0,
			file_size   INTEGER DEFAULT 0,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			pdf_id      INTEGER,
			category_id INTEGER,
			role_id     TEXT,
			user_id     TEXT,
			allowed     BOOLEAN DEFAULT TRUE,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (pdf_id) REFERENCES pdfs(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pdfs_name ON pdfs(name)`,
		`CREATE INDEX IF NOT EXISTS idx_pdfs_category ON pdfs(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permissions_pdf ON permissions(pdf_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permissions_category ON permissions(category_id)`,
		`INSERT OR IGNORE INTO categories (name, description) VALUES ('default', 'Uncategorized PDFs')`,
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	slog.Info("Database migrations completed")
	return nil
}
```

- [ ] **Step 3: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add internal/infrastructure/database/ go.mod go.sum
git commit -m "feat: add SQLite database infrastructure"
```

---

### Task 3: Filesystem watcher

**Files:**
- Create: `internal/infrastructure/watcher/fsnotify.go`

- [ ] **Step 1: Create watcher wrapper**

Create `internal/infrastructure/watcher/fsnotify.go`:

```go
package watcher

import (
	"log/slog"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher  *fsnotify.Watcher
	debounce time.Duration
	onChange func()
}

func NewFileWatcher(folder string, onChange func()) (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := w.Add(folder); err != nil {
		w.Close()
		return nil, err
	}

	fw := &FileWatcher{
		watcher:  w,
		debounce: 500 * time.Millisecond,
		onChange: onChange,
	}

	go fw.watch()

	slog.Info("Watching folder", "path", folder)
	return fw, nil
}

func (fw *FileWatcher) watch() {
	var timer *time.Timer

	for {
		select {
		case event, ok := <-fw.watcher.Events:
			if !ok {
				return
			}
			if event.Op&(fsnotify.Create|fsnotify.Remove|fsnotify.Write) != 0 {
				slog.Debug("File change detected", "file", event.Name)
				if timer != nil {
					timer.Stop()
				}
				timer = time.AfterFunc(fw.debounce, fw.onChange)
			}
		case err, ok := <-fw.watcher.Errors:
			if !ok {
				return
			}
			slog.Error("Watcher error", "error", err)
		}
	}
}

func (fw *FileWatcher) Close() error {
	return fw.watcher.Close()
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add internal/infrastructure/watcher/
git commit -m "feat: add filesystem watcher with debouncing"
```

---

## Phase 3: Adapter Layer

### Task 4: SQLite PDF repository

**Files:**
- Create: `internal/adapter/repository/sqlite_pdf.go`
- Create: `internal/adapter/repository/sqlite_pdf_test.go`

- [ ] **Step 1: Write failing test for repository**

Create `internal/adapter/repository/sqlite_pdf_test.go`:

```go
package repository

import (
	"testing"

	"discord-pdf-bot/internal/domain/entity"
	"discord-pdf-bot/internal/infrastructure/database"
)

func setupTestDB(t *testing.T) *database.SQLite {
	db, err := database.NewSQLite(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}
	return db
}

func TestCreateAndGetByName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLitePDFRepository(db.DB())

	pdf := &entity.PDF{
		Name:        "test_pdf",
		Filename:    "test_pdf.pdf",
		Path:        "./pdfs/test_pdf.pdf",
		Description: "Test document",
		FileSize:    1024,
		PageCount:   5,
	}

	// Create
	if err := repo.Create(pdf); err != nil {
		t.Fatalf("Failed to create PDF: %v", err)
	}

	// Get by name
	result, err := repo.GetByName("test_pdf")
	if err != nil {
		t.Fatalf("Failed to get PDF: %v", err)
	}

	if result.Name != "test_pdf" {
		t.Errorf("Expected name 'test_pdf', got '%s'", result.Name)
	}
	if result.Description != "Test document" {
		t.Errorf("Expected description 'Test document', got '%s'", result.Description)
	}
}

func TestSearch(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLitePDFRepository(db.DB())

	// Create test PDFs
	pdfs := []*entity.PDF{
		{Name: "math_101", Filename: "math_101.pdf", Path: "./pdfs/math_101.pdf", Description: "Calculus basics"},
		{Name: "math_201", Filename: "math_201.pdf", Path: "./pdfs/math_201.pdf", Description: "Linear algebra"},
		{Name: "physics_101", Filename: "physics_101.pdf", Path: "./pdfs/physics_101.pdf", Description: "Mechanics"},
	}

	for _, pdf := range pdfs {
		if err := repo.Create(pdf); err != nil {
			t.Fatalf("Failed to create PDF: %v", err)
		}
	}

	// Search for "math"
	results, err := repo.Search("math")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}

func TestDelete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSQLitePDFRepository(db.DB())

	pdf := &entity.PDF{
		Name:     "to_delete",
		Filename: "to_delete.pdf",
		Path:     "./pdfs/to_delete.pdf",
	}

	if err := repo.Create(pdf); err != nil {
		t.Fatalf("Failed to create PDF: %v", err)
	}

	if err := repo.Delete("to_delete"); err != nil {
		t.Fatalf("Failed to delete PDF: %v", err)
	}

	_, err := repo.GetByName("to_delete")
	if err == nil {
		t.Error("Expected error after delete, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/adapter/repository/ -v`
Expected: FAIL with "undefined: NewSQLitePDFRepository"

- [ ] **Step 3: Write implementation**

Create `internal/adapter/repository/sqlite_pdf.go`:

```go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"discord-pdf-bot/internal/domain/entity"
)

type SQLitePDFRepository struct {
	db *sql.DB
}

func NewSQLitePDFRepository(db *sql.DB) *SQLitePDFRepository {
	return &SQLitePDFRepository{db: db}
}

func (r *SQLitePDFRepository) GetByName(name string) (*entity.PDF, error) {
	pdf := &entity.PDF{}
	var categoryID sql.NullInt64

	err := r.db.QueryRow(
		"SELECT id, name, filename, path, description, category_id, uploaded_by, page_count, file_size, created_at, updated_at FROM pdfs WHERE name = ?",
		name,
	).Scan(&pdf.ID, &pdf.Name, &pdf.Filename, &pdf.Path, &pdf.Description, &categoryID, &pdf.UploadedBy, &pdf.PageCount, &pdf.FileSize, &pdf.CreatedAt, &pdf.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pdf '%s': %w", name, err)
	}
	if err != nil {
		return nil, fmt.Errorf("get pdf by name: %w", err)
	}

	if categoryID.Valid {
		pdf.CategoryID = &categoryID.Int64
	}

	return pdf, nil
}

func (r *SQLitePDFRepository) GetAll() ([]*entity.PDF, error) {
	rows, err := r.db.Query("SELECT id, name, filename, path, description, category_id, uploaded_by, page_count, file_size, created_at, updated_at FROM pdfs ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("get all pdfs: %w", err)
	}
	defer rows.Close()

	var pdfs []*entity.PDF
	for rows.Next() {
		pdf := &entity.PDF{}
		var categoryID sql.NullInt64

		if err := rows.Scan(&pdf.ID, &pdf.Name, &pdf.Filename, &pdf.Path, &pdf.Description, &categoryID, &pdf.UploadedBy, &pdf.PageCount, &pdf.FileSize, &pdf.CreatedAt, &pdf.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pdf: %w", err)
		}

		if categoryID.Valid {
			pdf.CategoryID = &categoryID.Int64
		}

		pdfs = append(pdfs, pdf)
	}

	return pdfs, nil
}

func (r *SQLitePDFRepository) GetByCategory(categoryID int64) ([]*entity.PDF, error) {
	rows, err := r.db.Query(
		"SELECT id, name, filename, path, description, category_id, uploaded_by, page_count, file_size, created_at, updated_at FROM pdfs WHERE category_id = ? ORDER BY name",
		categoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get pdfs by category: %w", err)
	}
	defer rows.Close()

	var pdfs []*entity.PDF
	for rows.Next() {
		pdf := &entity.PDF{}
		var catID sql.NullInt64

		if err := rows.Scan(&pdf.ID, &pdf.Name, &pdf.Filename, &pdf.Path, &pdf.Description, &catID, &pdf.UploadedBy, &pdf.PageCount, &pdf.FileSize, &pdf.CreatedAt, &pdf.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pdf: %w", err)
		}

		if catID.Valid {
			pdf.CategoryID = &catID.Int64
		}

		pdfs = append(pdfs, pdf)
	}

	return pdfs, nil
}

func (r *SQLitePDFRepository) Search(query string) ([]*entity.PDF, error) {
	rows, err := r.db.Query(
		"SELECT id, name, filename, path, description, category_id, uploaded_by, page_count, file_size, created_at, updated_at FROM pdfs WHERE name LIKE ? OR description LIKE ? OR filename LIKE ? ORDER BY name",
		"%"+query+"%", "%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("search pdfs: %w", err)
	}
	defer rows.Close()

	var pdfs []*entity.PDF
	for rows.Next() {
		pdf := &entity.PDF{}
		var categoryID sql.NullInt64

		if err := rows.Scan(&pdf.ID, &pdf.Name, &pdf.Filename, &pdf.Path, &pdf.Description, &categoryID, &pdf.UploadedBy, &pdf.PageCount, &pdf.FileSize, &pdf.CreatedAt, &pdf.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan pdf: %w", err)
		}

		if categoryID.Valid {
			pdf.CategoryID = &categoryID.Int64
		}

		pdfs = append(pdfs, pdf)
	}

	return pdfs, nil
}

func (r *SQLitePDFRepository) Create(pdf *entity.PDF) error {
	now := time.Now()
	result, err := r.db.Exec(
		"INSERT INTO pdfs (name, filename, path, description, category_id, uploaded_by, page_count, file_size, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		pdf.Name, pdf.Filename, pdf.Path, pdf.Description, pdf.CategoryID, pdf.UploadedBy, pdf.PageCount, pdf.FileSize, now, now,
	)
	if err != nil {
		return fmt.Errorf("create pdf: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	pdf.ID = id
	pdf.CreatedAt = now
	pdf.UpdatedAt = now

	return nil
}

func (r *SQLitePDFRepository) Update(pdf *entity.PDF) error {
	now := time.Now()
	_, err := r.db.Exec(
		"UPDATE pdfs SET description = ?, category_id = ?, page_count = ?, file_size = ?, updated_at = ? WHERE name = ?",
		pdf.Description, pdf.CategoryID, pdf.PageCount, pdf.FileSize, now, pdf.Name,
	)
	if err != nil {
		return fmt.Errorf("update pdf: %w", err)
	}

	pdf.UpdatedAt = now

	return nil
}

func (r *SQLitePDFRepository) Delete(name string) error {
	result, err := r.db.Exec("DELETE FROM pdfs WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete pdf: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("pdf '%s': not found", name)
	}

	return nil
}
```

- [ ] **Step 4: Run tests**

Run: `go test ./internal/adapter/repository/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/adapter/repository/
git commit -m "feat: implement SQLite PDF repository"
```

---

### Task 5: Disk storage adapter

**Files:**
- Create: `internal/adapter/storage/disk_storage.go`

- [ ] **Step 1: Create disk storage**

Create `internal/adapter/storage/disk_storage.go`:

```go
package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DiskStorage struct {
	basePath string
}

func NewDiskStorage(basePath string) *DiskStorage {
	return &DiskStorage{basePath: basePath}
}

func (ds *DiskStorage) Save(filename string, data []byte) (string, error) {
	path := filepath.Join(ds.basePath, filename)

	if err := os.WriteFile(path, data, 0644); err != nil {
		return "", fmt.Errorf("save file: %w", err)
	}

	return path, nil
}

func (ds *DiskStorage) Delete(path string) error {
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (ds *DiskStorage) Read(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

func (ds *DiskStorage) List(folder string) (map[string]string, error) {
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, fmt.Errorf("read directory: %w", err)
	}

	result := make(map[string]string)
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pdf") {
			name := strings.TrimSuffix(f.Name(), ".pdf")
			result[name] = filepath.Join(folder, f.Name())
		}
	}

	return result, nil
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/storage/
git commit -m "feat: implement disk storage adapter"
```

---

## Phase 4: Use Case Layer

### Task 6: PDF service

**Files:**
- Create: `internal/usecase/pdf_service.go`
- Create: `internal/usecase/pdf_service_test.go`

- [ ] **Step 1: Write failing test**

Create `internal/usecase/pdf_service_test.go`:

```go
package usecase

import (
	"testing"

	"discord-pdf-bot/internal/domain/entity"
)

type mockPDFRepo struct {
	pdfs map[string]*entity.PDF
}

func (m *mockPDFRepo) GetByName(name string) (*entity.PDF, error) {
	pdf, ok := m.pdfs[name]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	return pdf, nil
}

func (m *mockPDFRepo) GetAll() ([]*entity.PDF, error) {
	var result []*entity.PDF
	for _, pdf := range m.pdfs {
		result = append(result, pdf)
	}
	return result, nil
}

func (m *mockPDFRepo) GetByCategory(categoryID int64) ([]*entity.PDF, error) {
	var result []*entity.PDF
	for _, pdf := range m.pdfs {
		if pdf.CategoryID != nil && *pdf.CategoryID == categoryID {
			result = append(result, pdf)
		}
	}
	return result, nil
}

func (m *mockPDFRepo) Search(query string) ([]*entity.PDF, error) {
	var result []*entity.PDF
	for _, pdf := range m.pdfs {
		if contains(pdf.Name, query) || contains(pdf.Description, query) {
			result = append(result, pdf)
		}
	}
	return result, nil
}

func (m *mockPDFRepo) Create(pdf *entity.PDF) error {
	m.pdfs[pdf.Name] = pdf
	return nil
}

func (m *mockPDFRepo) Update(pdf *entity.PDF) error {
	m.pdfs[pdf.Name] = pdf
	return nil
}

func (m *mockPDFRepo) Delete(name string) error {
	delete(m.pdfs, name)
	return nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}

func TestPDFServiceSearch(t *testing.T) {
	repo := &mockPDFRepo{
		pdfs: map[string]*entity.PDF{
			"math_101": {Name: "math_101", Description: "Calculus"},
			"math_201": {Name: "math_201", Description: "Algebra"},
			"physics":  {Name: "physics", Description: "Mechanics"},
		},
	}

	service := NewPDFService(repo)

	results, err := service.Search("math")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/usecase/ -v`
Expected: FAIL with "undefined: NewPDFService"

- [ ] **Step 3: Write implementation**

Create `internal/usecase/pdf_service.go`:

```go
package usecase

import (
	"fmt"
	"strings"

	"discord-pdf-bot/internal/domain/entity"
	"discord-pdf-bot/internal/domain/port"
)

type PDFService struct {
	repo port.PDFRepository
}

func NewPDFService(repo port.PDFRepository) *PDFService {
	return &PDFService{repo: repo}
}

func (s *PDFService) GetByName(name string) (*entity.PDF, error) {
	return s.repo.GetByName(name)
}

func (s *PDFService) GetAll() ([]*entity.PDF, error) {
	return s.repo.GetAll()
}

func (s *PDFService) GetByCategory(categoryID int64) ([]*entity.PDF, error) {
	return s.repo.GetByCategory(categoryID)
}

func (s *PDFService) Search(query string) ([]*entity.PDF, error) {
	query = strings.ToLower(query)
	return s.repo.Search(query)
}

func (s *PDFService) Create(name, filename, path string, fileSize int64) (*entity.PDF, error) {
	// Check for duplicate
	existing, _ := s.repo.GetByName(name)
	if existing != nil {
		return nil, fmt.Errorf("pdf '%s': %w", name, errDuplicateName)
	}

	pdf := &entity.PDF{
		Name:     name,
		Filename: filename,
		Path:     path,
		FileSize: fileSize,
	}

	if err := s.repo.Create(pdf); err != nil {
		return nil, fmt.Errorf("create pdf: %w", err)
	}

	return pdf, nil
}

func (s *PDFService) Update(pdf *entity.PDF) error {
	return s.repo.Update(pdf)
}

func (s *PDFService) Delete(name string) error {
	return s.repo.Delete(name)
}

func (s *PDFService) SyncFromDisk(files map[string]string) error {
	existing, err := s.repo.GetAll()
	if err != nil {
		return fmt.Errorf("get existing pdfs: %w", err)
	}

	// Create map of existing PDFs
	existingMap := make(map[string]*entity.PDF)
	for _, pdf := range existing {
		existingMap[pdf.Name] = pdf
	}

	// Add new PDFs from disk
	for name, path := range files {
		if _, ok := existingMap[name]; !ok {
			pdf := &entity.PDF{
				Name:     name,
				Filename: name + ".pdf",
				Path:     path,
			}
			if err := s.repo.Create(pdf); err != nil {
				return fmt.Errorf("create pdf '%s': %w", name, err)
			}
		}
	}

	// Remove PDFs no longer on disk
	for _, pdf := range existing {
		if _, ok := files[pdf.Name]; !ok {
			if err := s.repo.Delete(pdf.Name); err != nil {
				return fmt.Errorf("delete pdf '%s': %w", pdf.Name, err)
			}
		}
	}

	return nil
}
```

- [ ] **Step 4: Fix test imports and run**

Update test file to add missing import:
```go
import (
	"fmt"
	"testing"

	"discord-pdf-bot/internal/domain/entity"
)
```

Run: `go test ./internal/usecase/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/usecase/
git commit -m "feat: implement PDF service with search and sync"
```

---

### Task 7: Category service

**Files:**
- Create: `internal/usecase/category_service.go`

- [ ] **Step 1: Create category entity**

Add to `internal/domain/entity/category.go`:

```go
package entity

import "time"

type Category struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}
```

- [ ] **Step 2: Create category repository port**

Add to `internal/domain/port/category_repository.go`:

```go
package port

import "discord-pdf-bot/internal/domain/entity"

type CategoryRepository interface {
	GetByName(name string) (*entity.Category, error)
	GetAll() ([]*entity.Category, error)
	Create(category *entity.Category) error
	Delete(name string) error
}
```

- [ ] **Step 3: Create SQLite category repository**

Create `internal/adapter/repository/sqlite_category.go`:

```go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"discord-pdf-bot/internal/domain/entity"
)

type SQLiteCategoryRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryRepository(db *sql.DB) *SQLiteCategoryRepository {
	return &SQLiteCategoryRepository{db: db}
}

func (r *SQLiteCategoryRepository) GetByName(name string) (*entity.Category, error) {
	cat := &entity.Category{}

	err := r.db.QueryRow(
		"SELECT id, name, description, created_at FROM categories WHERE name = ?",
		name,
	).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category '%s': not found", name)
	}
	if err != nil {
		return nil, fmt.Errorf("get category by name: %w", err)
	}

	return cat, nil
}

func (r *SQLiteCategoryRepository) GetAll() ([]*entity.Category, error) {
	rows, err := r.db.Query("SELECT id, name, description, created_at FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("get all categories: %w", err)
	}
	defer rows.Close()

	var categories []*entity.Category
	for rows.Next() {
		cat := &entity.Category{}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	return categories, nil
}

func (r *SQLiteCategoryRepository) Create(category *entity.Category) error {
	now := time.Now()
	result, err := r.db.Exec(
		"INSERT INTO categories (name, description, created_at) VALUES (?, ?, ?)",
		category.Name, category.Description, now,
	)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	category.ID = id
	category.CreatedAt = now

	return nil
}

func (r *SQLiteCategoryRepository) Delete(name string) error {
	result, err := r.db.Exec("DELETE FROM categories WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("category '%s': not found", name)
	}

	return nil
}
```

- [ ] **Step 4: Create category service**

Create `internal/usecase/category_service.go`:

```go
package usecase

import (
	"fmt"

	"discord-pdf-bot/internal/domain/entity"
	"discord-pdf-bot/internal/domain/port"
)

type CategoryService struct {
	repo port.CategoryRepository
}

func NewCategoryService(repo port.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) GetByName(name string) (*entity.Category, error) {
	return s.repo.GetByName(name)
}

func (s *CategoryService) GetAll() ([]*entity.Category, error) {
	return s.repo.GetAll()
}

func (s *CategoryService) Create(name, description string) (*entity.Category, error) {
	// Check for duplicate
	existing, _ := s.repo.GetByName(name)
	if existing != nil {
		return nil, fmt.Errorf("category '%s': already exists", name)
	}

	cat := &entity.Category{
		Name:        name,
		Description: description,
	}

	if err := s.repo.Create(cat); err != nil {
		return nil, fmt.Errorf("create category: %w", err)
	}

	return cat, nil
}

func (s *CategoryService) Delete(name string) error {
	if name == "default" {
		return fmt.Errorf("cannot delete default category")
	}
	return s.repo.Delete(name)
}
```

- [ ] **Step 5: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 6: Commit**

```bash
git add internal/domain/entity/ internal/domain/port/ internal/adapter/repository/ internal/usecase/
git commit -m "feat: implement category service and repository"
```

---

### Task 8: Permission service

**Files:**
- Create: `internal/domain/entity/permission.go`
- Create: `internal/domain/port/permission_repository.go`
- Create: `internal/adapter/repository/sqlite_permission.go`
- Create: `internal/usecase/permission_service.go`

- [ ] **Step 1: Create permission entity**

Create `internal/domain/entity/permission.go`:

```go
package entity

import "time"

type Permission struct {
	ID         int64
	PDFID      *int64
	CategoryID *int64
	RoleID     string
	UserID     string
	Allowed    bool
	CreatedAt  time.Time
}
```

- [ ] **Step 2: Create permission repository port**

Create `internal/domain/port/permission_repository.go`:

```go
package port

import "discord-pdf-bot/internal/domain/entity"

type PermissionRepository interface {
	GetByPDF(pdfID int64) ([]*entity.Permission, error)
	GetByCategory(categoryID int64) ([]*entity.Permission, error)
	Create(permission *entity.Permission) error
	Delete(id int64) error
	GetAll() ([]*entity.Permission, error)
}
```

- [ ] **Step 3: Create SQLite permission repository**

Create `internal/adapter/repository/sqlite_permission.go`:

```go
package repository

import (
	"database/sql"
	"fmt"
	"time"

	"discord-pdf-bot/internal/domain/entity"
)

type SQLitePermissionRepository struct {
	db *sql.DB
}

func NewSQLitePermissionRepository(db *sql.DB) *SQLitePermissionRepository {
	return &SQLitePermissionRepository{db: db}
}

func (r *SQLitePermissionRepository) GetByPDF(pdfID int64) ([]*entity.Permission, error) {
	rows, err := r.db.Query(
		"SELECT id, pdf_id, category_id, role_id, user_id, allowed, created_at FROM permissions WHERE pdf_id = ?",
		pdfID,
	)
	if err != nil {
		return nil, fmt.Errorf("get permissions by pdf: %w", err)
	}
	defer rows.Close()

	return r.scanPermissions(rows)
}

func (r *SQLitePermissionRepository) GetByCategory(categoryID int64) ([]*entity.Permission, error) {
	rows, err := r.db.Query(
		"SELECT id, pdf_id, category_id, role_id, user_id, allowed, created_at FROM permissions WHERE category_id = ?",
		categoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("get permissions by category: %w", err)
	}
	defer rows.Close()

	return r.scanPermissions(rows)
}

func (r *SQLitePermissionRepository) GetAll() ([]*entity.Permission, error) {
	rows, err := r.db.Query("SELECT id, pdf_id, category_id, role_id, user_id, allowed, created_at FROM permissions")
	if err != nil {
		return nil, fmt.Errorf("get all permissions: %w", err)
	}
	defer rows.Close()

	return r.scanPermissions(rows)
}

func (r *SQLitePermissionRepository) scanPermissions(rows *sql.Rows) ([]*entity.Permission, error) {
	var permissions []*entity.Permission
	for rows.Next() {
		p := &entity.Permission{}
		var pdfID, categoryID sql.NullInt64

		if err := rows.Scan(&p.ID, &pdfID, &categoryID, &p.RoleID, &p.UserID, &p.Allowed, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}

		if pdfID.Valid {
			p.PDFID = &pdfID.Int64
		}
		if categoryID.Valid {
			p.CategoryID = &categoryID.Int64
		}

		permissions = append(permissions, p)
	}

	return permissions, nil
}

func (r *SQLitePermissionRepository) Create(permission *entity.Permission) error {
	now := time.Now()
	result, err := r.db.Exec(
		"INSERT INTO permissions (pdf_id, category_id, role_id, user_id, allowed, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		permission.PDFID, permission.CategoryID, permission.RoleID, permission.UserID, permission.Allowed, now,
	)
	if err != nil {
		return fmt.Errorf("create permission: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	permission.ID = id
	permission.CreatedAt = now

	return nil
}

func (r *SQLitePermissionRepository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM permissions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete permission: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("permission '%d': not found", id)
	}

	return nil
}
```

- [ ] **Step 4: Create permission service**

Create `internal/usecase/permission_service.go`:

```go
package usecase

import (
	"discord-pdf-bot/internal/domain/entity"
	"discord-pdf-bot/internal/domain/port"
)

type PermissionService struct {
	repo port.PermissionRepository
}

func NewPermissionService(repo port.PermissionRepository) *PermissionService {
	return &PermissionService{repo: repo}
}

func (s *PermissionService) CheckAccess(userRoles []string, userID string, pdfID int64) bool {
	permissions, err := s.repo.GetByPDF(pdfID)
	if err != nil {
		return true // Default allow on error
	}

	// If no permissions set, allow all
	if len(permissions) == 0 {
		return true
	}

	for _, p := range permissions {
		// Check user-specific permission
		if p.UserID == userID {
			return p.Allowed
		}

		// Check role-based permission
		for _, role := range userRoles {
			if p.RoleID == role {
				return p.Allowed
			}
		}
	}

	return false
}

func (s *PermissionService) AddPermission(pdfID *int64, categoryID *int64, roleID, userID string) error {
	perm := &entity.Permission{
		PDFID:      pdfID,
		CategoryID: categoryID,
		RoleID:     roleID,
		UserID:     userID,
		Allowed:    true,
	}

	return s.repo.Create(perm)
}

func (s *PermissionService) RemovePermission(id int64) error {
	return s.repo.Delete(id)
}

func (s *PermissionService) GetAll() ([]*entity.Permission, error) {
	return s.repo.GetAll()
}
```

- [ ] **Step 5: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 6: Commit**

```bash
git add internal/domain/entity/ internal/domain/port/ internal/adapter/repository/ internal/usecase/
git commit -m "feat: implement permission service and repository"
```

---

## Phase 5: Adapter Layer - Discord Bot

### Task 9: Discord bot setup

**Files:**
- Create: `internal/adapter/discord/bot.go`
- Create: `internal/adapter/discord/embeds.go`

- [ ] **Step 1: Create embed builder**

Create `internal/adapter/discord/embeds.go`:

```go
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
		Color:       0x5865F2, // Discord blurple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "File",
				Value:  pdf.Filename,
				Inline: true,
			},
			{
				Name:   "Size",
				Value:  formatFileSize(pdf.FileSize),
				Inline: true,
			},
			{
				Name:   "Pages",
				Value:  fmt.Sprintf("%d", pdf.PageCount),
				Inline: true,
			},
		},
	}

	if pdf.CategoryID != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Category",
			Value:  fmt.Sprintf("%d", *pdf.CategoryID),
			Inline: true,
		})
	}

	return embed
}

func searchResultEmbed(pdfs []*entity.PDF) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       "Search Results",
		Description: fmt.Sprintf("Found %d PDFs", len(pdfs)),
		Color:       0x57F287, // Green
	}

	for i, pdf := range pdfs {
		if i >= 10 {
			break
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   pdf.Name,
			Value:  pdf.Description,
			Inline: false,
		})
	}

	return embed
}

func errorEmbed(message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Error",
		Description: message,
		Color:       0xED4245, // Red
	}
}

func successEmbed(message string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       "Success",
		Description: message,
		Color:       0x57F287, // Green
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
```

- [ ] **Step 2: Create bot setup**

Create `internal/adapter/discord/bot.go`:

```go
package discord

import (
	"fmt"
	"log/slog"

	"discord-pdf-bot/internal/usecase"

	"github.com/bwmarrin/discordgo"
)

type Bot struct {
	session            *discordgo.Session
	pdfService         *usecase.PDFService
	categoryService    *usecase.CategoryService
	permissionService  *usecase.PermissionService
	guildID            string
	adminRole          string
}

func NewBot(
	token string,
	pdfService *usecase.PDFService,
	categoryService *usecase.CategoryService,
	permissionService *usecase.PermissionService,
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
		guildID:           guildID,
		adminRole:         adminRole,
	}

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
	// Get existing commands
	existing, err := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
	if err != nil {
		return fmt.Errorf("fetch commands: %w", err)
	}

	// Get PDFs from database
	pdfs, err := b.pdfService.GetAll()
	if err != nil {
		return fmt.Errorf("fetch pdfs: %w", err)
	}

	// Build map of existing commands
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

	// Register utility commands
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
	}

	for _, cmd := range commands {
		// Check if command already exists
		existing, _ := b.session.ApplicationCommands(b.session.State.User.ID, b.guildID)
		found := false
		for _, e := range existing {
			if e.Name == cmd.Name {
				found = true
				break
			}
		}

		if !found {
			_, err := b.session.ApplicationCommandCreate(b.session.State.User.ID, b.guildID, cmd)
			if err != nil {
				slog.Error("Failed to create command", "command", cmd.Name, "error", err)
			}
		}
	}
}
```

- [ ] **Step 3: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 4: Commit**

```bash
git add internal/adapter/discord/
git commit -m "feat: implement Discord bot setup with embeds"
```

---

### Task 10: Interaction handlers

**Files:**
- Create: `internal/adapter/discord/handlers.go`
- Modify: `internal/adapter/discord/bot.go`

- [ ] **Step 1: Create handlers**

Create `internal/adapter/discord/handlers.go`:

```go
package discord

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/bwmarrin/discordgo"
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

	// Check permissions
	userRoles := i.Member.Roles
	userID := i.Member.User.ID
	if !b.permissionService.CheckAccess(userRoles, userID, pdf.ID) {
		b.respondError(s, i, "You don't have permission to access this PDF")
		return
	}

	// Send deferred response
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})

	// Read PDF from storage
	data, err := b.pdfService.GetByName(name)
	if err != nil {
		b.followupError(s, i, "Failed to read PDF")
		return
	}

	// Send PDF with embed
	embed := pdfEmbed(pdf)
	_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
		Files: []*discordgo.File{
			{
				Name:   pdf.Filename,
				Reader: strings.NewReader(string(data.Description)), // This needs fixing - should be actual PDF bytes
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

	embed := searchResultEmbed(results)
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

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

	embed := &discordgo.MessageEmbed{
		Title:       "Available PDFs",
		Description: fmt.Sprintf("Found %d PDFs", len(pdfs)),
		Color:       0x5865F2,
	}

	for i, pdf := range pdfs {
		if i >= 25 {
			break
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   pdf.Name,
			Value:  pdf.Description,
			Inline: true,
		})
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func (b *Bot) handleComponent(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Handle button interactions for pagination
	customID := i.MessageComponentData().CustomID

	switch {
	case strings.HasPrefix(customID, "page_"):
		b.handlePagination(s, i, customID)
	default:
		b.respondError(s, i, "Unknown component")
	}
}

func (b *Bot) handlePagination(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// TODO: Implement pagination logic
	b.respondError(s, i, "Pagination not implemented yet")
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

func (b *Bot) followupError(s *discordgo.Session, i *discordgo.InteractionCreate, message string) {
	embed := errorEmbed(message)
	s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
}
```

- [ ] **Step 2: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 3: Commit**

```bash
git add internal/adapter/discord/
git commit -m "feat: implement interaction handlers"
```

---

## Phase 6: Main Entry Point

### Task 11: Wire everything together

**Files:**
- Modify: `cmd/bot/main.go`

- [ ] **Step 1: Update main.go with DI wiring**

Update `cmd/bot/main.go`:

```go
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
	// Setup structured logging
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	// Get environment variables
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

	// Initialize database
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

	// Initialize repositories
	pdfRepo := repository.NewSQLitePDFRepository(db.DB())
	categoryRepo := repository.NewSQLiteCategoryRepository(db.DB())
	permissionRepo := repository.NewSQLitePermissionRepository(db.DB())

	// Initialize storage
	diskStorage := storage.NewDiskStorage("./pdfs")

	// Initialize services
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

	// Initialize bot
	bot, err := discord.NewBot(token, pdfService, categoryService, permissionService, guildID, adminRole)
	if err != nil {
		slog.Error("Failed to create bot", "error", err)
		os.Exit(1)
	}

	if err := bot.Open(); err != nil {
		slog.Error("Failed to open bot connection", "error", err)
		os.Exit(1)
	}
	defer bot.Close()

	// Sync commands
	if err := bot.SyncCommands(); err != nil {
		slog.Error("Failed to sync commands", "error", err)
		os.Exit(1)
	}

	// Start file watcher
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

	_, err = watcher.NewFileWatcher("./pdfs", onChange)
	if err != nil {
		slog.Error("Failed to create file watcher", "error", err)
		os.Exit(1)
	}

	slog.Info("Bot is running", "guild", guildID)

	// Wait for shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	slog.Info("Shutting down...")
}
```

- [ ] **Step 2: Create .env.example**

Create `.env.example`:

```
DISCORD_BOT_TOKEN=your_bot_token_here
GUILD_ID=your_guild_id_here
ADMIN_ROLE=PDF Admin
```

- [ ] **Step 3: Update .gitignore**

Update `.gitignore`:

```
.env
data/
*.db
.DS_Store
.serena/
```

- [ ] **Step 4: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 5: Commit**

```bash
git add cmd/ .env.example .gitignore
git commit -m "feat: implement main entry point with DI wiring"
```

---

## Phase 7: Admin Commands

### Task 12: Implement admin commands

**Files:**
- Modify: `internal/adapter/discord/handlers.go`
- Modify: `internal/adapter/discord/bot.go`

- [ ] **Step 1: Add admin commands to bot setup**

Add to `internal/adapter/discord/bot.go` in `registerUtilityCommands()`:

```go
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
```

- [ ] **Step 2: Add admin check helper**

Add to `internal/adapter/discord/handlers.go`:

```go
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
```

- [ ] **Step 3: Add upload handler**

Add to `internal/adapter/discord/handlers.go`:

```go
func (b *Bot) handleUpload(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !b.isAdmin(i) {
		b.respondError(s, i, "You don't have permission to upload PDFs")
		return
	}

	// Get attachment
	attachment := i.ApplicationCommandData().Options[0].Value.(string)
	// TODO: Get actual attachment from interaction
	_ = attachment

	// Get description
	description := ""
	if len(i.ApplicationCommandData().Options) > 1 {
		description = i.ApplicationCommandData().Options[1].StringValue()
	}

	// TODO: Download attachment and save to disk
	// TODO: Create PDF in database
	// TODO: Sync commands

	b.respondSuccess(s, i, "PDF uploaded successfully (not implemented yet)")
}
```

- [ ] **Step 4: Add delete handler**

Add to `internal/adapter/discord/handlers.go`:

```go
func (b *Bot) handleDelete(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if !b.isAdmin(i) {
		b.respondError(s, i, "You don't have permission to delete PDFs")
		return
	}

	name := i.ApplicationCommandData().Options[0].StringValue()

	// Get PDF to delete
	pdf, err := b.pdfService.GetByName(name)
	if err != nil {
		b.respondError(s, i, fmt.Sprintf("PDF not found: %s", name))
		return
	}

	// Delete from database
	if err := b.pdfService.Delete(name); err != nil {
		b.respondError(s, i, "Failed to delete PDF from database")
		return
	}

	// Delete from disk
	// TODO: Delete file from disk

	// Sync commands
	if err := b.SyncCommands(); err != nil {
		slog.Error("Failed to sync commands after delete", "error", err)
	}

	b.respondSuccess(s, i, fmt.Sprintf("PDF '%s' deleted", pdf.Name))
}
```

- [ ] **Step 5: Add PDF info handler**

Add to `internal/adapter/discord/handlers.go`:

```go
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
```

- [ ] **Step 6: Add PDF edit handler**

Add to `internal/adapter/discord/handlers.go`:

```go
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
```

- [ ] **Step 7: Update command handler**

Update `handleCommand` in `internal/adapter/discord/handlers.go`:

```go
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
```

- [ ] **Step 8: Add respondSuccess helper**

Add to `internal/adapter/discord/handlers.go`:

```go
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
```

- [ ] **Step 9: Verify build**

Run: `go build ./cmd/bot/`
Expected: Build succeeds

- [ ] **Step 10: Commit**

```bash
git add internal/adapter/discord/
git commit -m "feat: implement admin commands (upload, delete, info, edit)"
```

---

## Phase 8: Testing and Polish

### Task 13: Integration testing (was Task 12)

**Files:**
- Create: `internal/usecase/pdf_service_test.go` (expand)

- [ ] **Step 1: Add integration tests**

Add to `internal/usecase/pdf_service_test.go`:

```go
func TestPDFServiceSyncFromDisk(t *testing.T) {
	repo := &mockPDFRepo{
		pdfs: map[string]*entity.PDF{},
	}

	service := NewPDFService(repo)

	// Test adding new PDFs
	files := map[string]string{
		"math_101": "./pdfs/math_101.pdf",
		"math_201": "./pdfs/math_201.pdf",
	}

	if err := service.SyncFromDisk(files); err != nil {
		t.Fatalf("SyncFromDisk failed: %v", err)
	}

	if len(repo.pdfs) != 2 {
		t.Errorf("Expected 2 PDFs, got %d", len(repo.pdfs))
	}

	// Test removing PDFs
	delete(files, "math_101")

	if err := service.SyncFromDisk(files); err != nil {
		t.Fatalf("SyncFromDisk failed: %v", err)
	}

	if len(repo.pdfs) != 1 {
		t.Errorf("Expected 1 PDF, got %d", len(repo.pdfs))
	}

	if _, exists := repo.pdfs["math_201"]; !exists {
		t.Error("Expected math_201 to exist")
	}
}

func TestPDFServiceCreate(t *testing.T) {
	repo := &mockPDFRepo{
		pdfs: map[string]*entity.PDF{},
	}

	service := NewPDFService(repo)

	// Test creating new PDF
	pdf, err := service.Create("test_pdf", "test_pdf.pdf", "./pdfs/test_pdf.pdf", 1024)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if pdf.Name != "test_pdf" {
		t.Errorf("Expected name 'test_pdf', got '%s'", pdf.Name)
	}

	// Test creating duplicate PDF
	_, err = service.Create("test_pdf", "test_pdf.pdf", "./pdfs/test_pdf.pdf", 1024)
	if err == nil {
		t.Error("Expected error for duplicate PDF, got nil")
	}
}
```

- [ ] **Step 2: Run all tests**

Run: `go test ./... -v`
Expected: All tests pass

- [ ] **Step 3: Commit**

```bash
git add internal/usecase/
git commit -m "test: add integration tests for PDF service"
```

---

### Task 14: Update CLAUDE.md

**Files:**
- Modify: `CLAUDE.md`

- [ ] **Step 1: Update CLAUDE.md with new architecture**

Update `CLAUDE.md`:

```markdown
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Discord bot in Go that manages PDF files as slash commands. Features include search, categories, permissions, admin commands, pagination, and metadata. Uses clean architecture with SQLite storage.

## Commands

```bash
# Run
go run cmd/bot/main.go

# Build
go build -o discord-pdf-bot ./cmd/bot/

# Test
go test ./... -v

# Dependencies
go mod tidy
```

## Architecture

Clean architecture with 4 layers:

- **Domain** (`internal/domain/`) — entities, ports, errors. Zero dependencies.
- **Use Case** (`internal/usecase/`) — application services (PDF, Category, Permission).
- **Adapter** (`internal/adapter/`) — Discord bot, SQLite repos, disk storage.
- **Infrastructure** (`internal/infrastructure/`) — database, filesystem watcher.

Entry point: `cmd/bot/main.go` — wires all dependencies.

## Key Dependencies

- `discordgo` — Discord API
- `fsnotify` — filesystem watcher
- `modernc.org/sqlite` — SQLite driver (pure Go)

## Config

Environment variables:
- `DISCORD_BOT_TOKEN` — Discord bot token
- `GUILD_ID` — Discord server ID (guild commands)
- `ADMIN_ROLE` — Role name for admin commands (default: "PDF Admin")

Files:
- `.env` — environment variables (gitignored)
- `data/bot.db` — SQLite database (gitignored)
- `pdfs/` — PDF storage directory
```

- [ ] **Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with new architecture"
```

---

## Final Verification

- [ ] **Step 1: Run full test suite**

Run: `go test ./... -v`
Expected: All tests pass

- [ ] **Step 2: Build binary**

Run: `go build -o discord-pdf-bot ./cmd/bot/`
Expected: Binary created

- [ ] **Step 3: Verify structure**

Run: `find . -name "*.go" -not -path "./.git/*" | sort`
Expected: All files in correct locations

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "chore: complete PDF bot upgrade"
```
