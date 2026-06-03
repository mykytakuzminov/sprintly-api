package domain

import "errors"

var (
	ErrNotFound     = errors.New("resource not found")
	ErrForbidden    = errors.New("access denied")
	ErrConflict     = errors.New("resource already exists")
	ErrUnauthorized = errors.New("missing or invalid token")
	ErrBadRequest   = errors.New("invalid request body")
)
