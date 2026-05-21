package port

import "discord-pdf-bot/internal/domain/entity"

type PermissionRepository interface {
	GetByPDF(pdfID int64) ([]*entity.Permission, error)
	GetByCategory(categoryID int64) ([]*entity.Permission, error)
	Create(permission *entity.Permission) error
	Delete(id int64) error
	GetAll() ([]*entity.Permission, error)
}
