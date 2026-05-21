package entity

import "time"

type Category struct {
	ID          int64
	Name        string
	Description string
	CreatedAt   time.Time
}
