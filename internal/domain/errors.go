package domain

import "errors"

var (
	ErrPDFNotFound      = errors.New("pdf not found")
	ErrDuplicateName    = errors.New("pdf name already exists")
	ErrPermissionDenied = errors.New("permission denied")
	ErrInvalidInput     = errors.New("invalid input")
)
