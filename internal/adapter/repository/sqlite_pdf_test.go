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

	if err := repo.Create(pdf); err != nil {
		t.Fatalf("Failed to create PDF: %v", err)
	}

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
