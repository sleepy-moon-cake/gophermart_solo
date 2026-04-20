package shared

import "errors"

var (
	// db errors
	ErrWriteConflict = errors.New("resource conflict")
	ErrNotFound      = errors.New("resource not found")
	ErrAlreadyExists = errors.New("resource already exists")
	// service errors
	ErrNotMatchPassword = errors.New("not match")

	ErrUnauthorized = errors.New("not unauthorized")
)
