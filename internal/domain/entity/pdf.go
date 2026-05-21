package entity

import "time"

type PDF struct {
	ID          int64
	Name        string
	Filename    string
	Path        string
	Description string
	CategoryID  *int64
	UploadedBy  string
	PageCount   int
	FileSize    int64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
