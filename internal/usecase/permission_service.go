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
		return false // Fail closed on error
	}

	if len(permissions) == 0 {
		return true
	}

	for _, p := range permissions {
		if p.UserID == userID {
			return p.Allowed
		}

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
