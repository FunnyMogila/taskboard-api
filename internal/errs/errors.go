package errs

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrAlreadyExists     = errors.New("already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrProjectClosed     = errors.New("project is closed")
	ErrInvalidTransition = errors.New("invalid task status transition")
	ErrForbidden         = errors.New("forbidden")
)
