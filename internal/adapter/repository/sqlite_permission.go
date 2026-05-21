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

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate permissions: %w", err)
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
