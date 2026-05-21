package entity

import "time"

type Permission struct {
	ID         int64
	PDFID      *int64
	CategoryID *int64
	RoleID     string
	UserID     string
	Allowed    bool
	CreatedAt  time.Time
}
