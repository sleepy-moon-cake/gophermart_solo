package shared

import "errors"

var (
	// db errors
	ErrUserConflict = errors.New("user already exist")
	ErrNotFound = errors.New("user not found")
	// service errors 
	 ErrNotMatchPassword = errors.New("not match")
)