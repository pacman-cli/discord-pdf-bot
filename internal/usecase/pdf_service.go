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
	existing, _ := s.repo.GetByName(name)
	if existing != nil {
		return nil, fmt.Errorf("pdf '%s': already exists", name)
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

	existingMap := make(map[string]*entity.PDF)
	for _, pdf := range existing {
		existingMap[pdf.Name] = pdf
	}

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

	for _, pdf := range existing {
		if _, ok := files[pdf.Name]; !ok {
			if err := s.repo.Delete(pdf.Name); err != nil {
				return fmt.Errorf("delete pdf '%s': %w", pdf.Name, err)
			}
		}
	}

	return nil
}
