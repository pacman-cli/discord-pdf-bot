package repository

import (
	"database/sql"
	"fmt"
	"time"

	"discord-pdf-bot/internal/domain/entity"
)

type SQLiteCategoryRepository struct {
	db *sql.DB
}

func NewSQLiteCategoryRepository(db *sql.DB) *SQLiteCategoryRepository {
	return &SQLiteCategoryRepository{db: db}
}

func (r *SQLiteCategoryRepository) GetByName(name string) (*entity.Category, error) {
	cat := &entity.Category{}

	err := r.db.QueryRow(
		"SELECT id, name, description, created_at FROM categories WHERE name = ?",
		name,
	).Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("category '%s': not found", name)
	}
	if err != nil {
		return nil, fmt.Errorf("get category by name: %w", err)
	}

	return cat, nil
}

func (r *SQLiteCategoryRepository) GetAll() ([]*entity.Category, error) {
	rows, err := r.db.Query("SELECT id, name, description, created_at FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("get all categories: %w", err)
	}
	defer rows.Close()

	var categories []*entity.Category
	for rows.Next() {
		cat := &entity.Category{}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}

func (r *SQLiteCategoryRepository) Create(category *entity.Category) error {
	now := time.Now()
	result, err := r.db.Exec(
		"INSERT INTO categories (name, description, created_at) VALUES (?, ?, ?)",
		category.Name, category.Description, now,
	)
	if err != nil {
		return fmt.Errorf("create category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	category.ID = id
	category.CreatedAt = now

	return nil
}

func (r *SQLiteCategoryRepository) Delete(name string) error {
	result, err := r.db.Exec("DELETE FROM categories WHERE name = ?", name)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("category '%s': not found", name)
	}

	return nil
}
