package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	_ "modernc.org/sqlite"
)

type SQLite struct {
	db *sql.DB
}

func NewSQLite(dbPath string) (*SQLite, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	slog.Info("Connected to SQLite database", "path", dbPath)

	return &SQLite{db: db}, nil
}

func (s *SQLite) DB() *sql.DB {
	return s.db
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) Migrate() error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS categories (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL UNIQUE,
			description TEXT DEFAULT '',
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS pdfs (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			name        TEXT NOT NULL UNIQUE,
			filename    TEXT NOT NULL,
			path        TEXT NOT NULL,
			description TEXT DEFAULT '',
			category_id INTEGER,
			uploaded_by TEXT,
			page_count  INTEGER DEFAULT 0,
			file_size   INTEGER DEFAULT 0,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,
		`CREATE TABLE IF NOT EXISTS permissions (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			pdf_id      INTEGER,
			category_id INTEGER,
			role_id     TEXT,
			user_id     TEXT,
			allowed     BOOLEAN DEFAULT TRUE,
			created_at  DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (pdf_id) REFERENCES pdfs(id),
			FOREIGN KEY (category_id) REFERENCES categories(id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_pdfs_name ON pdfs(name)`,
		`CREATE INDEX IF NOT EXISTS idx_pdfs_category ON pdfs(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permissions_pdf ON permissions(pdf_id)`,
		`CREATE INDEX IF NOT EXISTS idx_permissions_category ON permissions(category_id)`,
		`INSERT OR IGNORE INTO categories (name, description) VALUES ('default', 'Uncategorized PDFs')`,
	}

	for _, m := range migrations {
		if _, err := s.db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	slog.Info("Database migrations completed")
	return nil
}
