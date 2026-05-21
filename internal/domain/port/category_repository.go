package port

import "discord-pdf-bot/internal/domain/entity"

type CategoryRepository interface {
	GetByName(name string) (*entity.Category, error)
	GetAll() ([]*entity.Category, error)
	Create(category *entity.Category) error
	Delete(name string) error
}
