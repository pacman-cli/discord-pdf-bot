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
