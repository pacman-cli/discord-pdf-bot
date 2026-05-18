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
