package usecase

import (
	"fmt"
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

func TestPDFServiceSyncFromDisk(t *testing.T) {
	repo := &mockPDFRepo{
		pdfs: map[string]*entity.PDF{},
	}

	service := NewPDFService(repo)

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

	pdf, err := service.Create("test_pdf", "test_pdf.pdf", "./pdfs/test_pdf.pdf", 1024)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if pdf.Name != "test_pdf" {
		t.Errorf("Expected name 'test_pdf', got '%s'", pdf.Name)
	}

	_, err = service.Create("test_pdf", "test_pdf.pdf", "./pdfs/test_pdf.pdf", 1024)
	if err == nil {
		t.Error("Expected error for duplicate PDF, got nil")
	}
}
