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
